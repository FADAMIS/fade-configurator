package ui

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/navidys/tvxwidgets"
	"github.com/rivo/tview"

	"go.bug.st/serial"
)

var state = State{
	ShowHiddenFiles: false,
	FirmwarePath:    "",
}

func createFilePicker() *tview.TreeView {
	homeDir, _ := os.UserHomeDir()
	root := tview.NewTreeNode(".")

	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	add := func(target *tview.TreeNode, path string) {
		files, err := os.ReadDir(path)
		if err != nil {
			panic(err)
		}

		for _, file := range files {
			if strings.HasPrefix(file.Name(), ".") && !state.ShowHiddenFiles {
				continue
			}

			node := tview.NewTreeNode(file.Name()).
				SetReference(filepath.Join(path, file.Name())).
				SetSelectable(true)

			if file.IsDir() {
				node.SetColor(tcell.ColorGreen)
			} else {
				node.SetColor(tcell.ColorWhite)
			}

			target.AddChild(node)
		}
	}

	add(root, homeDir)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return
		}

		path := reference.(string)

		// directories have green color
		if node.GetColor() == tcell.ColorGreen {
			children := node.GetChildren()
			if len(children) == 0 {
				add(node, path)
			} else {
				node.SetExpanded(!node.IsExpanded())
			}
		} else if node.GetColor() == tcell.ColorWhite { // files have white color
			state.FirmwarePath = path
		}
	})

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case rune('h'):
			state.ShowHiddenFiles = !state.ShowHiddenFiles

			tree.GetRoot().ClearChildren()

			/*childNodes := tree.GetRoot().GetChildren()
			if !ShowHidden {
				for _, node := range childNodes {
					if strings.HasPrefix(node.GetText(), ".") {
						tree.GetRoot().ClearChildren()
					}
				}
			}*/

			add(root, homeDir)
		}

		return event
	})

	return tree
}

func createPortSelector() *tview.DropDown {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}

	if len(ports) == 0 {
		ports = []string{"No serial port found"}
	}

	dropDown := tview.NewDropDown().
		SetLabel("Select port").
		SetOptions(ports, nil)

	return dropDown
}

func createGyroView() *tvxwidgets.BarChart {
	chart := tvxwidgets.NewBarChart()

	chart.AddBar("Yaw", 0, tcell.ColorMediumVioletRed)
	chart.AddBar("Pitch", 0, tcell.ColorDeepSkyBlue)
	chart.AddBar("Roll", 0, tcell.ColorLightYellow)

	chart.SetMaxValue(360)

	chart.SetAxesColor(tcell.ColorAntiqueWhite)
	chart.SetAxesLabelColor(tcell.ColorAntiqueWhite)

	return chart
}

func createPidFlex() *tview.Flex {
	p := tview.NewInputField()
	i := tview.NewInputField()
	d := tview.NewInputField()

	p.SetLabel("Input P value: ").
		SetFieldWidth(10).
		SetAcceptanceFunc(tview.InputFieldFloat)

	i.SetLabel("Input I value: ").
		SetFieldWidth(10).
		SetAcceptanceFunc(tview.InputFieldFloat)

	d.SetLabel("Input D value: ").
		SetFieldWidth(10).
		SetAcceptanceFunc(tview.InputFieldFloat)

	pidFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	pidFlex.AddItem(p, 0, 1, true)
	pidFlex.AddItem(i, 0, 1, true)
	pidFlex.AddItem(d, 0, 1, true)

	pidFlex.SetTitle("Set PID values")
	pidFlex.SetBorder(true)

	return pidFlex
}
