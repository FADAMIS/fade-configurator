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

	defer closePort()

	filePicker := createFilePicker()
	//portSelector := createPortSelector()
	//gyroView := createGyroView()
	//pidFlex := createPidFlex()
	logView := tview.NewTextView()
	config.AppState.LogView = logView
	//connectButton := createConnectButton()
	flashButton := createFlashButton()

	mainFlex := tview.NewFlex()
	mainFlex.AddItem(filePicker, 0, 1, true)
	mainFlex.AddItem(flashButton, 0, 1, true)
	mainFlex.SetDirection(tview.FlexColumn).AddItem(logView, 0, 1, true)

	err := app.SetRoot(mainFlex, true).EnableMouse(true).Run()
	if err != nil {
		log.Fatal(err)
	}
}

func closePort() {
	if config.AppState.Port != nil {
		config.AppState.Port.Close()
	}
}
