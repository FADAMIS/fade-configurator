package ui

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FADAMIS/fade-configurator/config"
	"github.com/FADAMIS/fade-configurator/device/dfu"
	"github.com/FADAMIS/fade-configurator/device/fsp"
	"github.com/gdamore/tcell/v2"
	"github.com/navidys/tvxwidgets"
	"github.com/rivo/tview"

	"go.bug.st/serial"
)

func createFilePicker() *tview.TreeView {
	homeDir, _ := os.UserHomeDir()
	root := tview.NewTreeNode(".")

	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	add := func(target *tview.TreeNode, path string) {
		files, err := os.ReadDir(path)
		if err != nil {
			slog.Error(err.Error())
			panic(err)
		}

		for _, file := range files {
			if strings.HasPrefix(file.Name(), ".") && !config.AppState.ShowHiddenFiles {
				continue
			}

			node := tview.NewTreeNode(file.Name()).
				SetReference(filepath.Join(path, file.Name())).
				SetSelectable(true)

			if file.IsDir() {
				node.SetColor(tcell.ColorGreen)
			} else if strings.HasSuffix(strings.ToLower(file.Name()), ".bin") {
				node.SetColor(tcell.ColorViolet)
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
		} else if node.GetColor() == tcell.ColorViolet { // selectable files have violet color
			config.AppState.FirmwarePath = path
		}
	})

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case rune('h'):
			config.AppState.ShowHiddenFiles = !config.AppState.ShowHiddenFiles

			tree.GetRoot().ClearChildren()

			add(root, homeDir)
		}

		return event
	})

	tree.SetBorder(true)
	tree.SetTitle("Select firmware")

	return tree
}

func createPortSelector() *tview.DropDown {
	ports, err := serial.GetPortsList()
	if err != nil {
		slog.Error(err.Error())
		log.Fatal(err)
	}

	if len(ports) == 0 {
		ports = []string{config.PortNotFound}
	}

	dropDown := tview.NewDropDown().
		SetLabel("Select port").
		SetOptions(ports, nil)

	dropDown.SetSelectedFunc(func(text string, index int) {
		config.AppState.SelectedPortName = text
		config.AppState.SelectedPortIndex = index
	})

	dropDown.SetBorder(true)
	dropDown.SetTitle("Select device")

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

	chart.SetBorder(true)
	chart.SetTitle("Gyro information")

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

func createConnectButton() *tview.Button {
	connectButton := tview.NewButton("Connect to device").
		SetSelectedFunc(func() {
			if config.AppState.SelectedPortName == config.PortNotFound {
				config.AppState.LogView.SetLabel("Error")
				config.AppState.LogView.SetText("No port selected")
				return
			}

			if config.AppState.Port != nil {
				config.AppState.Port.Close()
			}

			var err error

			config.AppState.Port, err = fsp.NewSerialDevice(config.AppState.SelectedPortName)

			if err != nil {
				slog.Error(err.Error())
				config.AppState.LogView.SetLabel("Error")
				config.AppState.LogView.SetText("Could not open port")
				return
			}
		})

	return connectButton
}

func startFlashing(firmware []byte, button *tview.Button) {
	go func() {
		end := make(chan int)
		go activityBarUpdate(end)

		err := config.AppState.DFU.FlashFirmware(firmware, func(status string) {
			config.AppState.App.QueueUpdateDraw(func() {
				config.AppState.LogView.SetLabel("Status")
				config.AppState.LogView.SetText(status)
			})
		})

		config.AppState.App.QueueUpdateDraw(func() {
			if err == nil {
				config.AppState.LogView.SetLabel("Status")
				config.AppState.LogView.SetText("Finished flashing")
			} else {
				slog.Error(err.Error())
				config.AppState.LogView.SetLabel("Error")
				config.AppState.LogView.SetText(err.Error())
			}

			button.SetDisabled(false)
		})

		end <- 0
	}()
}

func createFlashButton() *tview.Button {
	flashButton := tview.NewButton("Flash firmware")
	flashButton.SetSelectedFunc(func() {
		dev, err := dfu.NewDFUDevice()
		if err != nil {
			slog.Error(err.Error())
			config.AppState.LogView.SetLabel("Error")
			config.AppState.LogView.SetText("Could not open device")
			return
		}

		config.AppState.DFU = dev

		if config.AppState.FirmwarePath == "" {
			config.AppState.LogView.SetLabel("Error")
			config.AppState.LogView.SetText("No firmware selected")
			return
		}

		fw, err := os.ReadFile(config.AppState.FirmwarePath)
		if err != nil {
			slog.Error(err.Error())
			config.AppState.LogView.SetLabel("Error")
			config.AppState.LogView.SetText("Could not open file")
			return
		}

		flashButton.SetDisabled(true)

		startFlashing(fw, flashButton)
	})

	return flashButton
}

func createActivityBar() *tvxwidgets.ActivityModeGauge {
	bar := tvxwidgets.NewActivityModeGauge()
	bar.SetPgBgColor(tcell.ColorDarkViolet)
	bar.SetBorder(true)
	bar.SetTitle("Status")

	return bar
}

func activityBarUpdate(end chan int) {
	tick := time.NewTicker(50 * time.Millisecond)

	for {
		select {
		case <-end:
			tick.Stop()
			config.AppState.App.QueueUpdateDraw(func() {
				config.AppState.ActivityBar.Reset()
			})
			return

		case <-tick.C:
			config.AppState.App.QueueUpdateDraw(func() {
				config.AppState.ActivityBar.Pulse()
			})
		}
	}
}

func createLogoFlex() *tview.Flex {
	logoView := tview.NewTextView().SetText(config.Logo).SetTextAlign(tview.AlignCenter)
	description := tview.NewTextView().SetText(config.Description).SetTextAlign(tview.AlignCenter)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	flex.AddItem(nil, 0, 1, false)
	flex.AddItem(logoView, 0, 1, false)
	flex.AddItem(description, 0, 1, false)

	return flex
}
