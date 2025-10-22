package config

import (
	"github.com/FADAMIS/fade-configurator/device"
	"github.com/rivo/tview"
	"go.bug.st/serial"
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
)

type State struct {
	App               *tview.Application
	ShowHiddenFiles   bool
	FirmwarePath      string
	SelectedPortName  string
	SelectedPortIndex int
	Port              serial.Port
	DFU               *device.DFUDevice
	LogView           *tview.TextView
}

var AppState = State{
	App:               nil,
	ShowHiddenFiles:   false,
	FirmwarePath:      "",
	SelectedPortName:  PortNotFound,
	SelectedPortIndex: -1,
	DFU:               nil,
	Port:              nil,
	LogView:           tview.NewTextView(),
}
