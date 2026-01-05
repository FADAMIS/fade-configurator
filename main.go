package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/FADAMIS/fade-configurator/ui"
)

func main() {
	userConfig, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	logFilePath := filepath.Join(userConfig, "fade-configurator", "logs")
	err = os.MkdirAll(logFilePath, 0755)
	if err != nil {
		panic(err)
	}

	file, err := os.OpenFile(filepath.Join(logFilePath, "log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	logger := slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)

	ui.CreateApp()
}
