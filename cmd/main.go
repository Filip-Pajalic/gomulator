package main

import (
	"os"
	"app/internal/logger"
	"app/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		logger.Error("Usage: make <rom_file>")
	}
	romFile := os.Args[1]

	emuInstance := ui.StartEmulator(romFile)
	ui.UiInit(emuInstance)
}