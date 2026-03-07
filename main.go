package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"okemu/config"
	"okemu/logger"
	"okemu/okean240"
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

//go:embed hex/format.hex
var serialBytes []byte

//go:embed bin/zexall.com
var ramBytes []byte

var needReset = false

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

	w, raster, label := mainWindow(computer, conf)

	go emulator(computer)
	go screen(computer, raster, label, conf)
	(*w).ShowAndRun()
}

func mainWindow(computer *okean240.ComputerType, emuConfig *config.OkEmuConfig) (*fyne.Window, *canvas.Raster, *widget.Label) {
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
		widget.NewButton("RUN", func() {
			computer.SetRamBytes(ramBytes)
		}),
		widget.NewButton("DUMP", func() {
			computer.Dump(0x399, 15000)
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

func screen(computer *okean240.ComputerType, raster *canvas.Raster, label *widget.Label, emuConfig *config.OkEmuConfig) {
	ticker := time.NewTicker(20 * time.Millisecond)
	frame := 0
	var pre uint64 = 0
	var freq uint64 = 0

	for range ticker.C {
		if needReset {
			computer.Reset(emuConfig)
			needReset = false
		}
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

func emulator(computer *okean240.ComputerType) {
	ticker := time.NewTicker(133 * time.Nanosecond)
	var ticks = 0
	var ticksCPU = 0
	for range ticker.C {
		time.Sleep(133 * time.Nanosecond)
		ticks++
		if ticks%5 == 0 {
			// 1.5 MHz
			computer.TimerClk()
		}
		if ticks > ticksCPU {
			ticksCPU = ticks + computer.Do()*2
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
