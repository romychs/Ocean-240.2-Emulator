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
	"okemu/nanotime"
	"okemu/okean240"
	"okemu/z80/dis"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	//log "github.com/sirupsen/logrus"
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
	//	go timerClock(ctx, computer)
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
	var freq float64 = 0
	var freqTim float64 = 0
	timeStart := time.Now()

	for {
		select {
		case <-ticker.C:
			frame++
			// redraw screen here
			fyne.Do(func() {
				// status for every 50 frames
				if frame%50 == 0 {
					timeElapsed := time.Since(timeStart)
					period := float64(timeElapsed.Nanoseconds()) / 1_000_000.0

					freq = math.Round(float64(computer.Cycles()-pre)/period) / 1000.0
					freqTim = math.Round(float64(timerTicks.Load()-preTim)/period) / 1000.0
					label.SetText(formatLabel(computer, freq, freqTim))

					// adjust cpu clock
					if freq > 2.55 && cpuClkPeriod.Load() < defaultCpuClkPeriod+40 {
						cpuClkPeriod.Add(1)
					} else if freq < 2.45 && cpuClkPeriod.Load() > defaultCpuClkPeriod-40 {
						cpuClkPeriod.Add(-1)
					}
					// adjust timer clock
					if freqTim > 1.53 && timerClkPeriod.Load() < defaultTimerClkPeriod+20 {
						timerClkPeriod.Add(1)
					} else if freqTim < 1.47 && timerClkPeriod.Load() > defaultTimerClkPeriod-20 {
						timerClkPeriod.Add(-1)
					}

					//log.Debugf("Cpu clk period: %d, Timer clock period: %d, period: %1.3f", cpuClkPeriod.Load(), timerClkPeriod.Load(), period)
					pre = computer.Cycles()
					preTim = timerTicks.Load()
					timeStart = time.Now()
				}
				raster.Refresh()
			})
		case <-ctx.Done():
			return
		}
	}
}

func formatLabel(computer *okean240.ComputerType, freq float64, freqTim float64) string {
	return fmt.Sprintf("Screen size: %dx%d | Fcpu: %1.2fMHz | Ftmr: %1.2fMHz | Debugger: %s", computer.ScreenWidth(), computer.ScreenHeight(), freq, freqTim, computer.DebuggerState())
}

var timerTicks atomic.Uint64

const defaultTimerClkPeriod = 564 // = 1_000_000_000 / 1_607_900 // period in nanos for 1.5MHz frequency
const defaultCpuClkPeriod = 221   // = 1_000_000_000 / 2_770_000   // period in nanos for 2.5MHz frequency

var timerClkPeriod atomic.Int64 // = 1_000_000_000 / 1_607_900 // period in nanos for 1.5MHz frequency
var cpuClkPeriod atomic.Int64   // = 1_000_000_000 / 2_770_000   // period in nanos for 2.5MHz frequency

func emulator(ctx context.Context, computer *okean240.ComputerType) {
	ticker := time.NewTicker(133 * time.Nanosecond)
	defer ticker.Stop()

	cpuClkPeriod.Store(defaultCpuClkPeriod)
	timerClkPeriod.Store(defaultTimerClkPeriod)

	cpuTicks := uint64(0)
	nextTick := uint64(0)

	cpuTStart := nanotime.Now()
	tmrTStart := cpuTStart

	var bp uint16
	var bpType byte

	for {
		select {
		case <-ticker.C:
			tmrElapsed := nanotime.Since(tmrTStart)
			// TIMER CLK
			if tmrElapsed.Nanoseconds() >= timerClkPeriod.Load() {
				computer.TimerClk()
				timerTicks.Add(1)
				tmrTStart = nanotime.Now()
			}

			// CPU
			cpuElapsed := nanotime.Since(cpuTStart)
			if cpuElapsed.Nanoseconds() >= cpuClkPeriod.Load() {
				cpuTicks++
				bp = 0
				bpType = 0

				if fullSpeed.Load() {
					// Max frequency
					_, bp, bpType = computer.Do()
				} else {
					// 2.5MHz frequency
					if cpuTicks >= nextTick {
						var t uint32
						t, bp, bpType = computer.Do()
						nextTick += uint64(t)
					}
				}
				// Breakpoint hit
				if bp > 0 || bpType != 0 {
					listener.BreakpointHit(bp, bpType)
				}
				if needReset {
					computer.Reset()
					needReset = false
				}
				cpuTStart = nanotime.Now()
			}
		case <-ctx.Done():
			return
		}
	}

}
