package config

import (
	"github.com/FADAMIS/fade-configurator/device/dfu"
	"github.com/FADAMIS/fade-configurator/device/fsp"
	"github.com/navidys/tvxwidgets"
	"github.com/rivo/tview"
)

// Disables double borders on focused windows
func NormalizeBorders() {
	tview.Borders.HorizontalFocus = tview.BoxDrawingsLightHorizontal
	tview.Borders.VerticalFocus = tview.BoxDrawingsLightVertical
	tview.Borders.TopLeftFocus = tview.BoxDrawingsLightDownAndRight
	tview.Borders.TopRightFocus = tview.BoxDrawingsLightDownAndLeft
	tview.Borders.BottomLeftFocus = tview.BoxDrawingsLightUpAndRight
	tview.Borders.BottomRightFocus = tview.BoxDrawingsLightUpAndLeft
}

const (
	PortNotFound = "No serial port found"
	Logo         = `
    _________    ____  ______
   / ____/   |  / __ \/ ____/
  / /_  / /| | / / / / __/   
 / __/ / ___ |/ /_/ / /___   
/_/   /_/  |_/_____/_____/   `
	Description = "This is a configurator for FPV firmware FADE"
)

type State struct {
	App               *tview.Application
	ShowHiddenFiles   bool
	FirmwarePath      string
	SelectedPortName  string
	SelectedPortIndex int
	Port              *fsp.SerialDevice
	DFU               *dfu.DFUDevice
	LogView           *tview.TextView
	ActivityBar       *tvxwidgets.ActivityModeGauge
}

var AppState = State{
	App:               nil,
	ShowHiddenFiles:   false,
	FirmwarePath:      "",
	SelectedPortName:  PortNotFound,
	SelectedPortIndex: -1,
	DFU:               nil,
	Port:              nil,
	LogView:           nil,
	ActivityBar:       nil,
}
