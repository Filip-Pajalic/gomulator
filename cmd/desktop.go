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
	var dmgColors = flag.String("dmg-colors", "", "Enable GBC color palette for DMG games (optional: brown/green/grayscale/red)")
	flag.Parse()

	// Apply configuration
	if *debugMode {
		logger.Info("Debug mode enabled")
	}

	// Store DMG GBC color palette choice if flag is present
	// The actual palette will be applied after ROM is loaded in StartEmulator
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "dmg-colors" {
			// Flag was set, use the value or default to "auto" if empty
			palette := *dmgColors
			if palette == "" {
				palette = "auto"
			}
			ui.SetDMGColorsPaletteType(palette)
			logger.Info("DMG color palette mode set to: %s", palette)
		}
	})

	args := flag.Args()
	if len(args) < 1 {
		logger.Error("Usage: %s [options] <rom_file>", os.Args[0])
		logger.Info("Options:")
		logger.Info("  -debug              Enable debug mode")
		logger.Info("  -fps                Show FPS counter")
		logger.Info("  -dmg-colors[=type]  Enable GBC color palette for DMG games")
		logger.Info("                      Types: auto (default), green, brown, red, blue, grayscale")
		logger.Info("                      Note: 'auto' uses grayscale; game-specific detection not yet implemented")
		os.Exit(1)
	}

	romFile := args[0]

	emuInstance := ui.StartEmulator(romFile)
	ui.UiInit(emuInstance, *showFPS)
}
