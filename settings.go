package main

import (
	"okemu/config"
	"okemu/okean240/fdc"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	log "github.com/sirupsen/logrus"
)

func settingsDialog(config *config.OkEmuConfig) fyne.CanvasObject {

	debugger := widget.NewForm(
		widget.NewFormItem("Remote debug", debugCheckBox(config)),
		widget.NewFormItem("Host", hostEntry(config)),
		widget.NewFormItem("Port", portEntry(config)))

	cont := container.NewAppTabs(
		container.NewTabItemWithIcon("Debug", theme.MediaPlayIcon(), widget.NewCard("", "", debugger)),
	)
	for drv := byte(0); drv < fdc.TotalDrives; drv++ {
		cont.Append(
			container.NewTabItemWithIcon("Drive "+string(rune(66+drv))+":", theme.DocumentSaveIcon(), widget.NewCard("Floppy 720k", "", diskForm(config, drv))),
		)
	}
	return cont
}

func portEntry(cfg *config.OkEmuConfig) *widget.Entry {
	dbgPort := widget.NewEntry()
	dbgPort.SetText(strconv.Itoa(cfg.Debugger.Port))
	dbgPort.Validator = validation.NewRegexp(`^[0-9]+$`, "port can only contain numbers")
	dbgPort.OnSubmitted = func(s string) {
		p, e := strconv.Atoi(s)
		if e != nil {
			log.Warn("Illegal port number: " + s)
		} else {
			cfg.Debugger.Port = p
		}
	}
	return dbgPort
}

func hostEntry(cfg *config.OkEmuConfig) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetText(cfg.Debugger.Host)
	entry.Validator = validation.NewRegexp(`^[A-Za-z0-9_-]+$`, "hostname can only contain letters, numbers, '_', and '-'")
	entry.OnSubmitted = func(s string) {
		cfg.Debugger.Host = s
	}
	return entry
}

func debugCheckBox(cfg *config.OkEmuConfig) *widget.Check {
	// Debug Enabled
	check := widget.NewCheck("Enable", func(checked bool) {
		cfg.Debugger.Enabled = checked
	})
	check.Checked = cfg.Debugger.Enabled
	return check
}

func diskForm(cfg *config.OkEmuConfig, drv byte) *widget.Form {
	dskAutoLoad := widget.NewCheck("AutoLoad", func(checked bool) {
		cfg.FDC[drv].AutoLoad = checked
	})
	dskAutoLoad.Checked = cfg.FDC[drv].AutoLoad

	dskAutoSave := widget.NewCheck("AutoSave", func(checked bool) {
		cfg.FDC[drv].AutoSave = checked
	})

	dskAutoSave.Checked = cfg.FDC[drv].AutoSave

	dskFileName := widget.NewEntry()
	dskFileName.SetText(cfg.FDC[drv].FloppyFile)
	dskFileName.OnSubmitted = func(s string) {
		cfg.FDC[drv].FloppyFile = s
	}

	return widget.NewForm(
		widget.NewFormItem("AutoLoad", dskAutoLoad),
		widget.NewFormItem("AutoSave", dskAutoSave),
		widget.NewFormItem("File", dskFileName),
	)
}
