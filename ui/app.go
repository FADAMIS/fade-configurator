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
	logoView := tview.NewTextView().SetText(config.Logo)

	config.AppState.LogView = logView
	config.AppState.ActivityBar = activityBar

	activityFlex := tview.NewFlex()
	activityFlex.AddItem(activityBar, 0, 1, false)

	flashFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	flashFlex.AddItem(filePicker, 0, 15, true)
	flashFlex.AddItem(flashButton, 10, 1, true)

	serialFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	serialFlex.AddItem(portSelector, 0, 2, true)
	serialFlex.AddItem(pidFlex, 0, 2, true)
	serialFlex.AddItem(connectButton, 0, 1, true)

	infoFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	infoFlex.AddItem(logoView, 0, 2, false)
	infoFlex.AddItem(gyroView, 0, 2, false)
	infoFlex.AddItem(tview.NewBox(), 0, 1, false)

	componentsFlex := tview.NewFlex()
	componentsFlex.AddItem(infoFlex, 0, 2, true)
	componentsFlex.AddItem(serialFlex, 0, 2, true)
	componentsFlex.AddItem(flashFlex, 0, 1, true)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	mainFlex.AddItem(componentsFlex, 0, 10, true)
	mainFlex.AddItem(config.AppState.ActivityBar, 0, 1, true)

	err := app.SetRoot(mainFlex, true).EnableMouse(true).Run()
	if err != nil {
		slog.Error(err.Error())
		log.Fatal(err)
	}
}
