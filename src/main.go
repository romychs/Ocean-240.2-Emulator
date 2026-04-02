package main

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	"okemu/config"
	"okemu/debug"
	"okemu/debug/zrcp"
	"okemu/forms"
	"okemu/logger"
	"okemu/okean240"
	"runtime"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/loov/hrtime"
	"github.com/romychs/z80go/dis"
	log "github.com/sirupsen/logrus"
)

var Version = "v1.0.2"
var BuildTime = "2026-04-02"

const defaultTimerClkPeriod = 433
const defaultCpuClkPeriod = 310

const windowsTimerClkPeriod = 397
const windowsCpuClkPeriod = 298

const maxDelta = 5
const diffScale = 50.0

////go:embed hex/m80.hex
//var serialBytes []byte

////go:embed bin/jack.com
//var ramBytes []byte

var needReset = false

func main() {
	fmt.Printf("Starting Ocean-240.2 emulator %s build at %s\n", Version, BuildTime)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// base log init
	logger.InitLogging()

	// load config yml file
	config.LoadConfig()

	conf := config.GetConfig()

	// Reconfigure logging by config values
	logger.ReconfigureLogging(conf)

	if runtime.GOOS == "windows" {
		cpuClkPeriod.Store(windowsCpuClkPeriod)
		timerClkPeriod.Store(windowsTimerClkPeriod)
	} else {
		cpuClkPeriod.Store(defaultCpuClkPeriod)
		timerClkPeriod.Store(defaultTimerClkPeriod)
	}

	debugger := debug.NewDebugger()
	computer := okean240.NewComputer(conf, debugger)

	computer.AutoLoadFloppy()

	disassm := dis.NewDisassembler(computer)

	w, raster, label := forms.NewMainWindow(computer, conf, "Океан 240.2 "+Version)

	//dezog := dzrp.NewDZRP(conf, debugger, disassm, computer)
	dezog := zrcp.NewZRCP(conf, debugger, disassm, computer)

	go cpuClock(computer, dezog)
	go timerClock(computer)
	go screen(ctx, computer, raster, label)

	if conf.Debugger.Enabled {
		go dezog.SetupTcpHandler()
	}

	(*w).ShowAndRun()
	computer.AutoSaveFloppy()
	logger.CloseLogs()
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
					label.SetText(formatLabel(computer, cpuFreq, timerFreq, cpuClkPeriod.Load(), timerClkPeriod.Load()))

					adjustPeriods(computer, cpuFreq, timerFreq)

					log.Debugf("Cpu clk period: %d, Timer clock period: %d, frame time: %1.3fms", cpuClkPeriod.Load(), timerClkPeriod.Load(), period/50.0)
					logger.FlushLogs()
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

// adjustPeriods Adjust periods for CPU and Timer clock frequencies
func adjustPeriods(c *okean240.ComputerType, cpuFreq float64, timerFreq float64) {
	// adjust cpu clock if not full speed
	if !c.FullSpeed() {
		calcPeriod(cpuFreq, okean240.CPUFrequency, okean240.CPUFrequencyHi, okean240.CPUFrequencyLow, &cpuClkPeriod)
	}
	// adjust timerClock clock
	calcPeriod(timerFreq, okean240.TimerFrequency, okean240.TimerFrequencyHi, okean240.TimerFrequencyLow, &timerClkPeriod)
}

// calcPeriod  calc new value period to adjust frequency of timer or CPU
func calcPeriod(curFreq float64, destFreq float64, hiLimit float64, loLimit float64, period *atomic.Int64) {
	if curFreq > hiLimit && period.Load() < 2000 {
		period.Add(calcDelta(curFreq, destFreq))
	} else if curFreq < loLimit && period.Load() > 0 {
		period.Add(-calcDelta(curFreq, destFreq))
		if period.Load() < 0 {
			period.Store(0)
		}
	}
}

// calcDelta  calculate step to change period
func calcDelta(currentFreq float64, destFreq float64) int64 {
	delta := int64(math.Round(math.Abs(destFreq-currentFreq) * diffScale))
	if delta < 1 {
		return 1
	} else if delta > maxDelta {
		return maxDelta
	}
	return delta
}

func formatLabel(computer *okean240.ComputerType, freq float64, freqTim float64, cpu int64, tmr int64) string {
	return fmt.Sprintf("Screen size: %dx%d | Fcpu: %1.3fMHz | Ftmr: %1.3fMHz | Debugger: %s  CP:%d TP:%d",
		computer.ScreenWidth(), computer.ScreenHeight(), freq, freqTim, computer.DebuggerState(), cpu, tmr)
}

var timerTicks atomic.Uint64
var timerClkPeriod atomic.Int64 // period in nanos for 1.5MHz frequency
var cpuClkPeriod atomic.Int64   // period in nanos for 2.5MHz frequency

func timerClock(computer *okean240.ComputerType) {
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

func cpuClock(computer *okean240.ComputerType, dezog debug.DEZOG) {

	cpuTicks.Store(0)
	nextTick := uint64(0)

	var bp uint16
	var bpType byte
	timeStart := hrtime.Now()

	for {
		elapsed := hrtime.Since(timeStart)
		if int64(elapsed) >= cpuClkPeriod.Load() {
			timeStart = hrtime.Now()
			bp = 0
			bpType = 0

			// 2.5MHz frequency
			cpuTicks.Add(1)
			if computer.FullSpeed() {
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
				dezog.BreakpointHit(bp, bpType)
			}
			if needReset {
				computer.Reset()
				needReset = false
			}
		}
	}

}

//func initPerf() {
//	f, err := os.Create("okemu.prof")
//	if err != nil {
//		log.Warn("Can not create prof file", err)
//	}
//	defer func(f *os.File) {
//		err := f.Close()
//		if err != nil {
//			log.Warn("Can not close prof file", err)
//		}
//	}(f)
//	if err := pprof.StartCPUProfile(f); err != nil {
//		log.Warn("Can not start CPU profiling", err)
//	}
//	defer pprof.StopCPUProfile()
//}
