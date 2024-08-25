package main

import (
	"os"
	"pajalic.go.emulator/packages/emulator"
	log "pajalic.go.emulator/packages/logger"
)

func main() {
	if len(os.Args) < 2 {
		log.Error("Usage: make <rom_file>")
	}
	romFile := os.Args[1]

	emulator.StartEmulator(romFile)
}
