//go:build !js || !wasm

package main

import (
	"app/internal/logger"
	"app/internal/ui"
	"flag"
	"os"
)

func platformInit() {
	// Desktop-specific initialization
}

func platformMain() {
	// Parse command line flags
	var debugMode = flag.Bool("debug", false, "Enable debug mode")
	var showFPS = flag.Bool("fps", false, "Show FPS counter")
	flag.Parse()

	// Apply configuration
	if *debugMode {
		logger.Info("Debug mode enabled")
	}

	args := flag.Args()
	if len(args) < 1 {
		logger.Error("Usage: %s [options] <rom_file>", os.Args[0])
		logger.Info("Options:")
		logger.Info("  -debug        Enable debug mode")
		logger.Info("  -fps          Show FPS counter")
		os.Exit(1)
	}

	romFile := args[0]

	emuInstance := ui.StartEmulator(romFile)
	ui.UiInit(emuInstance, *showFPS)
}
