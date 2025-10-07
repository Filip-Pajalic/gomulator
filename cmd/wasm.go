//go:build js && wasm

package main

import (
	"app/internal/logger"
	"app/internal/ui"
	"syscall/js"
)

func platformInit() {
	// WASM-specific initialization
	logger.Info("Running in WASM/browser mode")
}

func platformMain() {
	logger.Info("Waiting for ROM from JavaScript...")

	js.Global().Set("startEmulatorWithROM", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			logger.Error("No ROM data provided from JS")
			return nil
		}

		romData := args[0]
		// Convert js.Value (Uint8Array) to Go []byte
		romBytes := make([]byte, romData.Get("length").Int())
		js.CopyBytesToGo(romBytes, romData)

		logger.Info("ROM received from JS (%d bytes), starting emulator...", len(romBytes))
		emuInstance := ui.StartEmulatorFromBytes(romBytes)

		// Run the UI in a goroutine to avoid blocking the JS event loop
		go ui.UiInit(emuInstance, false) // FPS disabled by default in WASM
		return nil
	}))

	// Keep Go WASM runtime alive
	select {}
}
