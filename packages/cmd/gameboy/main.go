package main

import (
	"os"
	"pajalic.go.emulator/packages/logger"
	"pajalic.go.emulator/packages/ui"
)

func main() {
	if len(os.Args) < 2 {
		logger.Error("Usage: make <rom_file>")
	}
	romFile := os.Args[1]

	emuInstance := ui.StartEmulator(romFile)
	ui.UiInit(emuInstance)
}
