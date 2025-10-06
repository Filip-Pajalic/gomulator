//go:build js && wasm

package main

import (
	"app/internal/logger"
	"app/internal/ui"
	"syscall/js"
)

func main() {
	logger.Info("Running in WASM/browser mode. Waiting for ROM from JS...")
	js.Global().Set("startEmulatorWithROM", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			logger.Error("No ROM data provided from JS")
			return nil
		}
		romData := args[0]
		// Convert js.Value (Uint8Array) to Go []byte
		romBytes := make([]byte, romData.Get("length").Int())
		js.CopyBytesToGo(romBytes, romData)
		
		logger.Info("ROM received from JS, starting emulator...")
		emuInstance := ui.StartEmulatorFromBytes(romBytes)
		// Run the UI in a goroutine to avoid blocking the JS event loop
		go ui.UiInit(emuInstance)
		return nil
	}))
	select {} // Keep Go WASM runtime alive
}
