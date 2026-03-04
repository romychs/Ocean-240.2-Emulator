package main

import (
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

func main() {

	fmt.Printf("Starting Ocean-240.2 emulator %s build at %s\n", Version, BuildTime)

	// base log init
	logger.InitLogging()

	// load config yml file
	config.LoadConfig()

	conf := config.GetConfig()

	// Reconfigure logging by config values
	//logger.ReconfigureLogging(conf)
	computer := okean240.New(conf)

	emulatorApp := app.New()
	w := emulatorApp.NewWindow("Океан 240.2")
	w.Canvas().SetOnTypedKey(
		func(key *fyne.KeyEvent) {
			computer.PutKey(key)
		})

	addShortcuts(w.Canvas(), computer)

	label := widget.NewLabel(fmt.Sprintf("Screen size: %dx%d", computer.ScreenWidth(), computer.ScreenHeight()))

	raster := canvas.NewRasterWithPixels(
		func(x, y, w, h int) color.Color {
			return computer.GetPixel(uint16(x/2), uint16(y/2))
		})
	raster.Resize(fyne.NewSize(512, 512))
	raster.SetMinSize(fyne.NewSize(512, 512))

	centerRaster := container.NewCenter(raster)

	w.Resize(fyne.NewSize(600, 600))

	hBox := container.NewHBox(
		widget.NewButton("Ctrl+C", func() {
			computer.PutCtrlKey(0x03)
		}),
		widget.NewSeparator(),
		widget.NewButton("Reset", func() {
			computer.Reset()
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

	go emulator(computer, raster, label)

	w.ShowAndRun()
}

func emulator(computer *okean240.ComputerType, raster *canvas.Raster, label *widget.Label) {
	ticker := time.NewTicker(133 * time.Nanosecond)
	var ticks = 0
	var ticksCPU = 3
	//var ticksSCR = TicksPerFrame
	//var frameStartTime = time.Now().UnixMicro()
	frameNextTime := time.Now().UnixMicro() + 20000
	frame := 0
	var pre uint64 = 0
	var freq uint64 = 0

	nextSecond := time.Now().Add(time.Second).UnixMicro()
	curScrWidth := 256
	for range ticker.C {
		ticks++
		if ticks%5 == 0 {
			// 1.5 MHz
			computer.TimerClk()
		}
		if ticks > ticksCPU {
			ticksCPU = ticks + computer.Do()*2
		}

		if time.Now().UnixMicro() > nextSecond {
			nextSecond = time.Now().Add(time.Second).UnixMicro()
			freq = computer.Cycles() - pre
			pre = computer.Cycles()
		}

		//if ticks >= ticksSCR {
		if time.Now().UnixMicro() > frameNextTime {
			frameNextTime = time.Now().UnixMicro() + 20000
			//ticksSCR = ticks + TicksPerFrame
			frame++
			// redraw screen here
			fyne.Do(func() {
				// check for screen mode changed
				if computer.ScreenWidth() != curScrWidth {
					curScrWidth = computer.ScreenWidth()
					newSize := fyne.NewSize(float32(curScrWidth*2), float32(computer.ScreenHeight()*2))
					raster.SetMinSize(newSize)
					raster.Resize(newSize)
				}
				// status for every 25 frames
				if frame%50 == 0 {
					label.SetText(fmt.Sprintf("Screen size: %dx%d  F: %d", computer.ScreenWidth(), computer.ScreenHeight(), freq))
				}
				raster.Refresh()
			})
		}
	}
}

func addShortcuts(c fyne.Canvas, computer *okean240.ComputerType) {
	// Add shortcuts for Ctrl+A to Ctrl+Z
	for kName := 'A'; kName <= 'Z'; kName++ {
		kk := fyne.KeyName(kName)
		sc := &desktop.CustomShortcut{KeyName: kk, Modifier: fyne.KeyModifierControl}
		c.AddShortcut(sc, func(shortcut fyne.Shortcut) { computer.PutCtrlKey(byte(kName&0xff) - 0x40) })
	}
}
