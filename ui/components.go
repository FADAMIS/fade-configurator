package ui

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
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
	dropDown := tview.NewDropDown().SetLabel("Select port")

	refreshDropDown(dropDown)

	dropDown.SetSelectedFunc(func(text string, index int) {
		slog.Info("selected text: " + text)
		config.AppState.SelectedPortName = text
		config.AppState.SelectedPortIndex = index
	})

	return dropDown
}

func portCallback(text string, index int) {
	slog.Info("selected text: " + text)
	config.AppState.SelectedPortName = text
	config.AppState.SelectedPortIndex = index
}

func refreshDropDown(dropdown *tview.DropDown) {
	ports, err := serial.GetPortsList()
	if err != nil {
		slog.Error(err.Error())
	}

	if len(ports) == 0 {
		ports = []string{config.PortNotFound}
	}

	dropdown.SetOptions(ports, portCallback)
}

func createRefreshButton(dropdown *tview.DropDown) *tview.Button {
	refresh := tview.NewButton("Refresh ports")
	refresh.SetSelectedFunc(func() {
		refreshDropDown(dropdown)
	})

	return refresh
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

func createConnectButton(pidFlex *tview.Flex) *tview.Button {
	connectButton := tview.NewButton("Connect to device").
		SetSelectedFunc(func() {
			go func() {
				if config.AppState.SelectedPortName == config.PortNotFound {
					slog.Error("No port selected")

					config.AppState.App.QueueUpdateDraw(func() {
						config.AppState.LogView.SetLabel("Error")
						config.AppState.LogView.SetText("No port selected")
					})
					return
				}

				if config.AppState.Port != nil {
					config.AppState.Port.Close()
				}

				var err error

				config.AppState.Port, err = fsp.NewSerialDevice(config.AppState.SelectedPortName)
				config.AppState.App.QueueUpdateDraw(func() { getPidValues(pidFlex) })

				if err != nil {
					slog.Error(err.Error())
					slog.Error("Could not open port")

					config.AppState.App.QueueUpdateDraw(func() {
						config.AppState.LogView.SetLabel("Error")
						config.AppState.LogView.SetText("Could not open port")
					})
					return
				}

				config.AppState.App.QueueUpdateDraw(func() {
					config.AppState.LogView.SetLabel("Info")
					config.AppState.LogView.SetText("Successfully connected")
				})
			}()
		})

	return connectButton
}

func getPidValues(pidFlex *tview.Flex) {
	p, _ := config.AppState.Port.GetValue(fsp.KEY_P_VALUE)
	i, _ := config.AppState.Port.GetValue(fsp.KEY_I_VALUE)
	d, _ := config.AppState.Port.GetValue(fsp.KEY_D_VALUE)

	pidFlex.GetItem(0).(*tview.InputField).SetText(strconv.FormatFloat(float64(p), 'f', -1, 32))
	pidFlex.GetItem(1).(*tview.InputField).SetText(strconv.FormatFloat(float64(i), 'f', -1, 32))
	pidFlex.GetItem(2).(*tview.InputField).SetText(strconv.FormatFloat(float64(d), 'f', -1, 32))
}

func createSaveButton(pidFlex *tview.Flex) *tview.Button {
	saveButton := tview.NewButton("Save values")
	saveButton.
		SetSelectedFunc(func() {
			go func() {
				if config.AppState.Port == nil {
					return
				}

				end := make(chan int)
				go activityBarUpdate(end)

				p, _ := strconv.ParseFloat(pidFlex.GetItem(0).(*tview.InputField).GetText(), 32)
				i, _ := strconv.ParseFloat(pidFlex.GetItem(1).(*tview.InputField).GetText(), 32)
				d, _ := strconv.ParseFloat(pidFlex.GetItem(2).(*tview.InputField).GetText(), 32)

				config.AppState.Port.SetValue(fsp.KEY_P_VALUE, float32(p))
				config.AppState.Port.SetValue(fsp.KEY_I_VALUE, float32(i))
				config.AppState.Port.SetValue(fsp.KEY_D_VALUE, float32(d))

				end <- 0
			}()

		})

	return saveButton
}

func startFlashing(firmware []byte, button *tview.Button) {
	go func() {
		end := make(chan int)
		go activityBarUpdate(end)

		err := config.AppState.DFU.FlashFirmware(firmware, func(status string) {
			config.AppState.App.QueueUpdateDraw(func() {
				slog.Info(status)
				config.AppState.LogView.SetLabel("Status")
				config.AppState.LogView.SetText(status)
			})
		})

		config.AppState.App.QueueUpdateDraw(func() {
			if err == nil {
				slog.Info("Finished flashing")
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
	bar.SetPgBgColor(tview.Styles.ContrastSecondaryTextColor)

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
