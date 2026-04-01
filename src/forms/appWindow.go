package forms

import (
	"fmt"
	"image/color"
	"okemu/config"
	"okemu/okean240"
	"okemu/okean240/fdc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewMainWindow(computer *okean240.ComputerType, config *config.OkEmuConfig) (*fyne.Window, *canvas.Raster, *widget.Label) {
	emulatorApp := app.New()
	w := emulatorApp.NewWindow("Океан 240.2")

	// Handle all keys typed in main window canvas
	w.Canvas().SetOnTypedKey(
		func(key *fyne.KeyEvent) {
			computer.PutKey(key)
		})

	w.Canvas().SetOnTypedRune(
		func(key rune) {
			computer.PutRune(key)
		})

	addShortcuts(w.Canvas(), computer)

	// ---
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

	//vBox := container.NewVBox(
	//	newToolbar(computer, w, emulatorApp, config),
	//	centerRaster,
	//	label,
	//)
	vBox := container.NewBorder(
		newToolbar(computer, w, emulatorApp, config),
		label,
		nil,
		nil,
		centerRaster,
	)
	w.SetContent(vBox)

	return &w, raster, label
}

func newToolbar(c *okean240.ComputerType, w fyne.Window, a fyne.App, config *config.OkEmuConfig) *fyne.Container {
	hBox := container.NewHBox()
	for d := 0; d < fdc.TotalDrives; d++ {
		hBox.Add(widget.NewLabel(string(rune(66+d)) + ":"))
		hBox.Add(widget.NewToolbar(
			widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
				err := c.SaveFloppy(byte(d))
				if err != nil {
					dialog.ShowError(err, w)
				}
			}),
			//widget.NewToolbarSpacer(),
			widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
				err := c.LoadFloppy(byte(d))
				if err != nil {
					dialog.ShowError(err, w)
				}
			}),
		))
		hBox.Add(widget.NewSeparator())
	}

	//hBox.Add(widget.NewButtonWithIcon("1", theme.DownloadIcon(), func() {
	//	c.SetRamBytes(ramBytes)
	//}))
	//
	//hBox.Add(widget.NewSeparator())
	hBox.Add(widget.NewButtonWithIcon("^C", theme.MediaStopIcon(), func() {
		c.PutCtrlKey(0x03)
	}))
	hBox.Add(widget.NewSeparator())
	bNorm := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), nil)
	bFast := widget.NewButtonWithIcon("", theme.MediaFastForwardIcon(), nil)

	bFast.OnTapped = func() {
		c.SetFullSpeed(true)
		c.SetCPUFrequency(50_000_000)
		bNorm.Enable()
		bFast.Disable()
	}

	bNorm.OnTapped = func() {
		c.SetFullSpeed(false)
		c.SetCPUFrequency(2_500_000)
		bNorm.Disable()
		bFast.Enable()
	}

	bNorm.Disable()

	//hBox.Add(cbFreq)
	hBox.Add(bNorm)
	hBox.Add(bFast)

	hBox.Add(widget.NewSeparator())
	hBox.Add(widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		cfg := config.Clone()
		d := dialog.NewCustomConfirm("OK-Emu settings", "Save", "Cancel", settingsDialog(cfg), func(b bool) {
			if b {
				cfg.Save()
			}
		}, w)
		d.Resize(fyne.NewSize(450, 360))
		//w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
		d.Show()
	}))
	hBox.Add(layout.NewSpacer())
	hBox.Add(widget.NewButtonWithIcon("Reset", theme.MediaReplayIcon(), func() {
		c.SetPendingReset(true)
		//computer.Reset(conf)
	}))
	return hBox
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
