package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"okemu/config"
	"okemu/debuger"
	"okemu/logger"
	"okemu/okean240"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
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

	computer := okean240.New(conf)
	computer.SetSerialBytes(serialBytes)
	computer.LoadFloppy()

	w, raster, label := mainWindow(computer)

	go emulator(computer)
	go screen(computer, raster, label)
	go debuger.SetupTcpHandler(conf, computer)

	(*w).ShowAndRun()
}

func mainWindow(computer *okean240.ComputerType) (*fyne.Window, *canvas.Raster, *widget.Label) {
	emulatorApp := app.New()
	w := emulatorApp.NewWindow("Океан 240.2")
	w.Canvas().SetOnTypedKey(
		func(key *fyne.KeyEvent) {
			computer.PutKey(key)
		})

	w.Canvas().SetOnTypedRune(
		func(key rune) {
			computer.PutRune(key)
		})

	addShortcuts(w.Canvas(), computer)

	label := widget.NewLabel(fmt.Sprintf("Screen size: %dx%d", computer.ScreenWidth(), computer.ScreenHeight()))

	raster := canvas.NewRasterWithPixels(
		func(x, y, w, h int) color.Color {
			var xx uint16
			if computer.ScreenWidth() == 512 {
				xx = uint16(x)
			} else {
				xx = uint16(x) / 2
			}
			return computer.GetPixel(xx, uint16(y/2))
		})
	raster.Resize(fyne.NewSize(512, 512))
	raster.SetMinSize(fyne.NewSize(512, 512))

	centerRaster := container.NewCenter(raster)

	w.Resize(fyne.NewSize(600, 600))

	hBox := container.NewHBox(
		//widget.NewButton("++", func() {
		//	computer.IncOffset()
		//}),
		//widget.NewButton("--", func() {
		//	computer.DecOffset()
		//}),
		widget.NewButton("Ctrl+C", func() {
			computer.PutCtrlKey(0x03)
		}),
		widget.NewButton("Load Floppy", func() {
			computer.LoadFloppy()
		}),
		widget.NewButton("Save Floppy", func() {
			computer.SaveFloppy()
		}),
		widget.NewButton("RUN1", func() {
			computer.SetRamBytes(ramBytes1)
		}),
		widget.NewButton("RUN2", func() {
			computer.SetRamBytes(ramBytes2)
		}),
		widget.NewButton("DUMP", func() {
			computer.Dump(0x100, 15000)
		}),
		widget.NewCheck("Full speed", func(b bool) {
			fullSpeed.Store(b)
			if b {
				computer.SetCPUFrequency(50_000_000)
			} else {
				computer.SetCPUFrequency(2_500_000)
			}
		}),
		widget.NewSeparator(),
		widget.NewButton("Reset", func() {
			needReset = true
			//computer.Reset(conf)
		}),
		widget.NewSeparator(),
		widget.NewButton("Закрыть", func() {
			emulatorApp.Quit()
		}),
	)
	vBox := container.NewVBox(
		centerRaster,
		label,
		hBox,
	)

	w.SetContent(vBox)

	return &w, raster, label
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
				label.SetText(fmt.Sprintf("Screen size: %dx%d  F: %d", computer.ScreenWidth(), computer.ScreenHeight(), freq))
			}
			raster.Refresh()
		})
	}
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
		if !computer.IsStepMode() || computer.IsRunMode() {
			var bp uint16 = 0
			if fullSpeed.Load() {
				_, bp = computer.Do()
			} else {
				if ticks >= nextClock {
					var t uint32
					t, bp = computer.Do()
					nextClock = ticks + uint64(t)*ticksPerTact
				}
			}
			// Breakpoint hit
			if bp > 0 {
				debuger.BreakpointHit(bp)
			}
		}
		if needReset {
			computer.Reset()
			needReset = false
		}
	}
}

// Add shortcuts for all Ctrl+<Letter>
func addShortcuts(c fyne.Canvas, computer *okean240.ComputerType) {
	// Add shortcuts for Ctrl+A to Ctrl+Z
	for kName := 'A'; kName <= 'Z'; kName++ {
		kk := fyne.KeyName(kName)
		sc := &desktop.CustomShortcut{KeyName: kk, Modifier: fyne.KeyModifierControl}
		c.AddShortcut(sc, func(shortcut fyne.Shortcut) { computer.PutCtrlKey(byte(kName&0xff) - 0x40) })
	}
}
