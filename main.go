package main

import (
	_ "embed"
	"fmt"
	"okemu/config"
	"okemu/debug"
	"okemu/debug/listener"
	"okemu/logger"
	"okemu/okean240"
	"okemu/okean240/fdc"
	"okemu/z80/dis"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

var Version = "v1.0.0"
var BuildTime = "2026-03-01"

//go:embed hex/m80.hex
var serialBytes []byte

//go:embed bin/main.com
var ramBytes1 []byte

//go:embed bin/PLOT.BAS
var ramBytes2 []byte

var needReset = false
var fullSpeed atomic.Bool

func main() {

	fmt.Printf("Starting Ocean-240.2 emulator %s build at %s\n", Version, BuildTime)

	// base log init
	logger.InitLogging()

	// load config yml file
	config.LoadConfig()

	conf := config.GetConfig()

	// Reconfigure logging by config values
	// logger.ReconfigureLogging(conf)

	debugger := debug.NewDebugger()
	computer := okean240.NewComputer(conf, debugger)

	computer.SetSerialBytes(serialBytes)

	if conf.FDC.AutoLoadB {
		err := computer.LoadFloppy(fdc.FloppyB)
		if err != nil {
			// show message
		}
	}
	if conf.FDC.AutoLoadC {
		err := computer.LoadFloppy(fdc.FloppyC)
		if err != nil {
			// show message
		}
	}

	disasm := dis.NewDisassembler(computer)

	w, raster, label := mainWindow(computer)

	go emulator(computer)
	go screen(computer, raster, label)

	if conf.Debugger.Enabled {
		go listener.SetupTcpHandler(conf, debugger, disasm, computer)
	}

	(*w).ShowAndRun()
}

func screen(computer *okean240.ComputerType, raster *canvas.Raster, label *widget.Label) {
	ticker := time.NewTicker(20 * time.Millisecond)
	frame := 0
	var pre uint64 = 0
	var freq uint64 = 0

	for range ticker.C {
		frame++
		// redraw screen here
		fyne.Do(func() {
			// status for every 50 frames
			if frame%50 == 0 {
				freq = computer.Cycles() - pre
				pre = computer.Cycles()
				label.SetText(formatLabel(computer, freq))
			}
			raster.Refresh()
		})
	}
}

func formatLabel(computer *okean240.ComputerType, freq uint64) string {
	return fmt.Sprintf("Screen size: %dx%d | F: %d | Debugger: %s", computer.ScreenWidth(), computer.ScreenHeight(), freq, computer.DebuggerState())
}

const ticksPerTact uint64 = 4

func emulator(computer *okean240.ComputerType) {
	ticker := time.NewTicker(66 * time.Nanosecond)
	var ticks uint64 = 0
	var nextClock = ticks + ticksPerTact
	//var ticksCPU = 3
	for range ticker.C {
		ticks++
		if ticks%10 == 0 {
			// 1.5 MHz
			computer.TimerClk()
		}
		var bp uint16 = 0
		var typ byte = 0
		if fullSpeed.Load() {
			_, bp, typ = computer.Do()
		} else {
			if ticks >= nextClock {
				var t uint32
				t, bp, typ = computer.Do()
				nextClock = ticks + uint64(t)*ticksPerTact
			}
		}
		// Breakpoint hit
		if bp > 0 || typ != 0 {
			listener.BreakpointHit(bp, typ)
		}
		if needReset {
			computer.Reset()
			needReset = false
		}
	}
}
