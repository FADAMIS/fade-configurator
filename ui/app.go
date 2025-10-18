package ui

import (
	"log"

	"github.com/rivo/tview"
)

func CreateApp() {
	NormalizeBorders()

	app := tview.NewApplication()

	mainFlex := tview.NewFlex()

	err := app.SetRoot(mainFlex, true).EnableMouse(true).Run()
	if err != nil {
		log.Fatal(err)
	}
}
