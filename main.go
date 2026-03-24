package main

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	"okemu/config"
	"okemu/debug"
	"okemu/debug/listener"
	"okemu/logger"
	"okemu/okean240"
	"okemu/z80/dis"
	"os"
	"runtime"
	"runtime/pprof"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/loov/hrtime"
	log "github.com/sirupsen/logrus"
)

var Version = "v1.0.0"
var BuildTime = "2026-03-01"

////go:embed hex/m80.hex
//var serialBytes []byte

//go:embed bin/2048.com
var ramBytes1 []byte

//go:embed bin/JACK.COM
var ramBytes2 []byte

var needReset = false

var fullSpeed atomic.Bool

func main() {

	fmt.Printf("Starting Ocean-240.2 emulator %s build at %s\n", Version, BuildTime)

	f, err := os.Create("okemu.prof")
	if err != nil {
		log.Warn("Can not create prof file", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Warn("Can not close prof file", err)
		}
	}(f)
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Warn("Can not start CPU profiling", err)
	}
	defer pprof.StopCPUProfile()

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// base log init
	logger.InitLogging()

	// load config yml file
	config.LoadConfig()

	conf := config.GetConfig()

	// Reconfigure logging by config values
	// logger.ReconfigureLogging(conf)

	debugger := debug.NewDebugger()
	computer := okean240.NewComputer(conf, debugger)

	//computer.SetSerialBytes(serialBytes)

	computer.AutoLoadFloppy()

	disasm := dis.NewDisassembler(computer)

	w, raster, label := mainWindow(computer, conf)

	go emulator(ctx, computer)
	go timerClock(ctx, computer)
	go screen(ctx, computer, raster, label)

	if conf.Debugger.Enabled {
		go listener.SetupTcpHandler(conf, debugger, disasm, computer)
	}

	(*w).ShowAndRun()
	computer.AutoSaveFloppy()
}

func screen(ctx context.Context, computer *okean240.ComputerType, raster *canvas.Raster, label *widget.Label) {
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	frame := 0
	var pre uint64 = 0
	var preTim uint64 = 0
	var cpuFreq float64 = 0
	var timerFreq float64 = 0
	timeStart := hrtime.Now()

	for {
		select {
		case <-ticker.C:
			frame++
			// redraw screen here
			fyne.Do(func() {
				// status for every 50 frames
				if frame%50 == 0 {
					timeElapsed := hrtime.Since(timeStart)
					period := float64(timeElapsed.Nanoseconds()) / 1_000_000.0

					//cpuFreq = math.Round(float64(computer.Cycles()-pre)/period) / 1000.0
					cpuFreq = math.Round(float64(cpuTicks.Load()-pre)/period) / 1000.0
					timerFreq = math.Round(float64(timerTicks.Load()-preTim)/period) / 1000.0
					label.SetText(formatLabel(computer, cpuFreq, timerFreq))

					adjustTimers(cpuFreq, timerFreq)

					//log.Debugf("Cpu clk period: %d, Timer clock period: %d, period: %1.3f", cpuClkPeriod.Load(), timerClkPeriod.Load(), period)
					//pre = computer.Cycles()
					pre = cpuTicks.Load()
					preTim = timerTicks.Load()
					timeStart = hrtime.Now()
				}
				raster.Refresh()
			})
		case <-ctx.Done():
			return
		}
	}
}

func adjustTimers(cpuFreq float64, timerFreq float64) {
	// adjust cpu clock
	if cpuFreq > 2.55 && cpuClkPeriod.Load() < defaultCpuClkPeriod+defaultCpuClkPeriod/2 {
		cpuClkPeriod.Add(1)
		//						cpuTicker.Reset(time.Duration(cpuClkPeriod.Load()))
	} else if cpuFreq < 2.45 && cpuClkPeriod.Load() > 3 {
		cpuClkPeriod.Add(-2)
		//						cpuTicker.Reset(time.Duration(cpuClkPeriod.Load()))
	}
	// adjust timerClock clock
	if timerFreq > 1.53 && timerClkPeriod.Load() < defaultTimerClkPeriod+defaultTimerClkPeriod/2 {
		timerClkPeriod.Add(1)
		//timerTicker.Reset(time.Duration(timerClkPeriod.Load()))
	} else if timerFreq < 1.47 && timerClkPeriod.Load() > 3 {
		timerClkPeriod.Add(-2)
		//timerTicker.Reset(time.Duration(timerClkPeriod.Load()))
	}
}

func formatLabel(computer *okean240.ComputerType, freq float64, freqTim float64) string {
	return fmt.Sprintf("Screen size: %dx%d | Fcpu: %1.3fMHz | Ftmr: %1.3fMHz | Debugger: %s", computer.ScreenWidth(), computer.ScreenHeight(), freq, freqTim, computer.DebuggerState())
}

var timerTicks atomic.Uint64

const defaultTimerClkPeriod = 430 // = 1_000_000_000 / 1_607_900 // period in nanos for 1.5MHz frequency
const defaultCpuClkPeriod = 311   // = 1_000_000_000 / 2_770_000   // period in nanos for 2.5MHz frequency

var timerClkPeriod atomic.Int64 // = 1_000_000_000 / 1_607_900 // period in nanos for 1.5MHz frequency
var cpuClkPeriod atomic.Int64   // = 1_000_000_000 / 2_770_000   // period in nanos for 2.5MHz frequency

//var timerTicker *time.Ticker

//var cpuTicker *time.Ticker

func timerClock(ctx context.Context, computer *okean240.ComputerType) {
	timerClkPeriod.Store(defaultTimerClkPeriod)
	timeStart := hrtime.Now()
	for {
		elapsed := hrtime.Since(timeStart)
		if int64(elapsed) > timerClkPeriod.Load() {
			timeStart = hrtime.Now()
			computer.TimerClk()
			timerTicks.Add(1)
			runtime.Gosched()
		}
	}

}

var cpuTicks atomic.Uint64

func emulator(ctx context.Context, computer *okean240.ComputerType) {
	cpuClkPeriod.Store(defaultCpuClkPeriod)
	//cpuTicker = time.NewTicker(time.Duration(cpuClkPeriod.Load()) * time.Nanosecond)
	//defer cpuTicker.Stop()

	cpuTicks.Store(0) // := uint64(0)
	nextTick := uint64(0)

	var bp uint16
	var bpType byte
	timeStart := hrtime.Now()

	for {
		//select {
		//case <-cpuTicker.C:
		// CPU
		elapsed := hrtime.Since(timeStart)
		if int64(elapsed) >= cpuClkPeriod.Load() {
			timeStart = hrtime.Now()
			bp = 0
			bpType = 0

			// 2.5MHz frequency
			cpuTicks.Add(1)
			if fullSpeed.Load() {
				// Max frequency
				_, bp, bpType = computer.Do()
			} else if cpuTicks.Load() >= nextTick {
				var t uint32
				t, bp, bpType = computer.Do()
				nextTick = cpuTicks.Load() + uint64(t)
				runtime.Gosched()
			}

			// Breakpoint hit
			if bp > 0 || bpType != 0 {
				listener.BreakpointHit(bp, bpType)
			}
			if needReset {
				computer.Reset()
				needReset = false
			}
		}
		//case <-ctx.Done():
		//	return
		//}
	}

}
