package ui

import (
	"log"

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
	logoView := tview.NewTextView().SetText(config.Logo)

	config.AppState.LogView = logView
	config.AppState.ActivityBar = activityBar

	flashFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	flashFlex.AddItem(filePicker, 0, 15, true)
	flashFlex.AddItem(flashButton, 10, 1, true)
	flashFlex.AddItem(activityBar, 0, 1, true)

	serialFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	serialFlex.AddItem(portSelector, 0, 2, true)
	serialFlex.AddItem(connectButton, 0, 1, true)
	serialFlex.AddItem(pidFlex, 0, 3, true)

	infoFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	infoFlex.AddItem(logoView, 0, 1, false)
	infoFlex.AddItem(gyroView, 0, 1, false)

	mainFlex := tview.NewFlex()
	mainFlex.AddItem(infoFlex, 0, 2, true)
	mainFlex.AddItem(serialFlex, 0, 2, true)
	mainFlex.AddItem(flashFlex, 0, 1, true)

	err := app.SetRoot(mainFlex, true).EnableMouse(true).Run()
	if err != nil {
		log.Fatal(err)
	}
}
