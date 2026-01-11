package ui

import (
	"log"
	"log/slog"

	"github.com/FADAMIS/fade-configurator/config"
	"github.com/rivo/tview"
)

func CreateApp() {
	config.NormalizeBorders()

	app := tview.NewApplication()
	config.AppState.App = app

	defer config.AppState.DFU.Close()
	defer config.AppState.Port.Close()

	filePicker := createFilePicker()
	portSelector := createPortSelector()
	gyroView := createGyroView()
	pidFlex := createPidFlex()
	logView := tview.NewTextView()
	connectButton := createConnectButton()
	flashButton := createFlashButton()
	activityBar := createActivityBar()
	logo := createLogoFlex()

	logView.SetTitle("Info")
	logView.SetBorder(true)

	config.AppState.LogView = logView
	config.AppState.ActivityBar = activityBar

	activityFlex := tview.NewFlex()
	activityFlex.AddItem(activityBar, 0, 1, false)

	flashFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	flashFlex.AddItem(filePicker, 0, 10, true)
	flashFlex.AddItem(flashButton, 0, 1, true)

	serialFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	serialFlex.AddItem(
		tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(pidFlex, 0, 9, true).
			AddItem(tview.NewButton("Save values"), 0, 1, true), 0, 1, true)

	serialFlex.AddItem(
		tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(portSelector, 0, 1, true).
			AddItem(gyroView, 0, 8, false).
			AddItem(connectButton, 0, 1, true), 0, 1, true)

	pages := tview.NewPages()
	pages.AddAndSwitchToPage("Main", logo, true)
	pages.AddPage("Config", serialFlex, true, false)
	pages.AddPage("Firmware", flashFlex, true, false)

	bar := tview.NewFlex().SetDirection(tview.FlexColumn)
	bar.SetBorder(true)
	bar.SetTitle("Navigation bar")

	makeButton := func(page string) *tview.Button {
		return tview.NewButton(page).SetSelectedFunc(func() {
			pages.SwitchToPage(page)
		})
	}

	bar.AddItem(makeButton("Config"), 0, 4, false)
	bar.AddItem(tview.NewBox(), 0, 1, false) // for padding, didnt find a better solution because button widget is acting weird
	bar.AddItem(makeButton("Firmware"), 0, 4, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	layout.AddItem(bar, 0, 1, true)
	layout.AddItem(pages, 0, 15, true)
	layout.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(config.AppState.ActivityBar, 0, 1, false).
		AddItem(config.AppState.LogView, 0, 1, false), 0, 1, true)

	err := app.SetRoot(layout, true).EnableMouse(true).Run()
	if err != nil {
		slog.Error(err.Error())
		log.Fatal(err)
	}
}
