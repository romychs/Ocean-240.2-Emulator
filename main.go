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
	"fyne.io/fyne/v2/widget"
)

var Version = "v1.0.0"
var BuildTime = "2026-03-01"

func main() {

	// base log init
	logger.InitLogging()

	// load config yml file
	config.LoadConfig()

	conf := config.GetConfig()

	// Reconfigure logging by config values
	//logger.ReconfigureLogging(conf)

	emuapp := app.New()
	w := emuapp.NewWindow("Океан 240.2")

	computer := okean240.New(conf)

	label := widget.NewLabel(fmt.Sprintf("Screen size: %dx%d", computer.ScreenWidth(), computer.ScreenHeight()))

	raster := canvas.NewRasterWithPixels(
		func(x, y, w, h int) color.Color {
			return computer.GetPixel(uint16(x/2), uint16(y/2))
		})
	raster.Resize(fyne.NewSize(512, 512))
	raster.SetMinSize(fyne.NewSize(512, 512))
	w.Resize(fyne.NewSize(600, 600))

	hBox := container.NewHBox(
		widget.NewButton("Reset", func() {
			computer.Reset()
		}),
		widget.NewButton("Закрыть", func() {
			emuapp.Quit()
		}),
	)
	vBox := container.NewVBox(
		raster,
		label,
		hBox,
	)
	w.SetContent(vBox)

	go emulator(computer, raster, label)

	w.ShowAndRun()

	//println("Tick computer")
	//computer := okean240.New(conf)
	//println("Run computer")
	//computer.Run()
}

const TicksPerFrame = 20_000_000 / 133

func emulator(computer *okean240.ComputerType, raster *canvas.Raster, label *widget.Label) {
	ticker := time.NewTicker(133 * time.Nanosecond)
	var ticks = 0
	var ticksCPU = 3
	//var ticksSCR = TicksPerFrame
	//var frameStartTime = time.Now().UnixMicro()
	frameNextTime := time.Now().UnixMicro() + 20000
	frame := 0
	curScrWidth := 256
	for range ticker.C {
		ticks++
		if ticks%5 == 0 {
			// 1.5 MHz
			computer.TimerClk()
		}
		if ticks > ticksCPU {
			ticksCPU = ticks + computer.Do()*3
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
				if frame%25 == 0 {
					label.SetText(fmt.Sprintf("Screen size: %dx%d  Tick: %d", computer.ScreenWidth(), computer.ScreenHeight(), computer.Cycles()))
				}
				raster.Refresh()
			})
		}
	}
}
