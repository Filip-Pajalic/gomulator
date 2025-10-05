package main

import (
	"app/internal/common"
	"app/internal/logger"
	"app/internal/ui"
	"flag"
	"os"
)

func main() {
	// Parse command line flags
	var skipBootAnim = flag.Bool("skip-boot", false, "Skip the Nintendo logo boot animation")
	var debugMode = flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	// Apply configuration
	if *skipBootAnim {
		common.SetBootAnimationEnabled(false)
		logger.Info("Boot animation disabled via command line")
	}

	if *debugMode {
		common.GlobalConfig.LogLevel = "DEBUG"
		logger.Info("Debug mode enabled")
	}

	args := flag.Args()
	if len(args) < 1 {
		logger.Error("Usage: %s [options] <rom_file>", os.Args[0])
		logger.Info("Options:")
		logger.Info("  -skip-boot    Skip the Nintendo logo boot animation")
		logger.Info("  -debug        Enable debug mode")
		os.Exit(1)
	}

	romFile := args[0]

	emuInstance := ui.StartEmulator(romFile)
	ui.UiInit(emuInstance)
}
