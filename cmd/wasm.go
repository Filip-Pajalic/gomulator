//go:build js && wasm

package main

import (
	"app/internal/input"
	"app/internal/logger"
	"app/internal/ui"
	"syscall/js"
)

// currentEmu holds the active emulator instance (if any) so debug JS functions
// can access the bus for ad-hoc reads.
var currentEmu *ui.EmuContext

// ROMStartConfig holds configuration for starting a ROM
type ROMStartConfig struct {
	ROMBytes   []byte
	ColorMode  string // "auto", "green", "grayscale", "brown", "red", "blue", or "" for default
}

func platformInit() {
	// WASM-specific initialization
	logger.Info("Running in WASM/browser mode")

	// Also log directly to JS console to verify logging works
	js.Global().Get("console").Call("log", "🔧 Go platformInit called - WASM is running!")
}

func platformMain() {
	logger.Info("Waiting for ROM from JavaScript...")
	// Channel used to send ROM configuration to the main goroutine so UiInit runs
	// on the main thread (required by some windowing/JS interactions).
	romStartCh := make(chan ROMStartConfig, 1)

	js.Global().Set("startEmulatorWithROM", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			logger.Error("No ROM data provided from JS")
			return nil
		}

		romData := args[0]
		// Convert js.Value (Uint8Array) to Go []byte
		romBytes := make([]byte, romData.Get("length").Int())
		js.CopyBytesToGo(romBytes, romData)

		// Optional second parameter: color mode
		colorMode := "auto" // Default to auto palette detection
		if len(args) >= 2 && !args[1].IsUndefined() && !args[1].IsNull() {
			colorMode = args[1].String()
			logger.Info("ROM: Color mode specified from JS: %s", colorMode)
		}

		logger.Info("ROM received from JS (%d bytes), enqueuing for start...", len(romBytes))
		// Enqueue the ROM configuration for the main goroutine to pick up and start the UI
		select {
		case romStartCh <- ROMStartConfig{ROMBytes: romBytes, ColorMode: colorMode}:
		default:
			// If channel already has a pending startup, drop or log
			logger.Warn("startEmulatorWithROM: previous ROM start pending, ignoring new request")
		}

		return nil
	}))

	// Expose a simple function emuInput(button, pressed) so host page can call directly
	emuInput := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 2 {
			return nil
		}
		btn := args[0].String()
		pressed := args[1].Bool()

		// Safety check: only process input if emulator is running
		if currentEmu == nil || !currentEmu.Running {
			js.Global().Get("console").Call("warn", "emuInput: emulator not running, ignoring input")
			return nil
		}

		st := input.GetState()
		if st == nil {
			js.Global().Get("console").Call("warn", "emuInput: input state not initialized")
			return nil
		}

		switch btn {
		case "up":
			st.Up = pressed
		case "down":
			st.Down = pressed
		case "left":
			st.Left = pressed
		case "right":
			st.Right = pressed
		case "a":
			st.A = pressed
		case "b":
			st.B = pressed
		case "start":
			st.Start = pressed
		case "select":
			st.Select = pressed
		default:
			js.Global().Get("console").Call("warn", "emuInput: unknown button", btn)
		}

		return nil
	})
	// Keep reference in global so it won't be garbage collected
	js.Global().Set("emuInput", emuInput)

	// Also listen for postMessage events (host can postMessage {type: 'emu-input', button, pressed})
	msgHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return nil
		}
		ev := args[0]
		data := ev.Get("data")
		if data.IsUndefined() || data.IsNull() {
			return nil
		}
		if data.Get("type").String() == "emu-input" {
			// Safety check: only process input if emulator is running
			if currentEmu == nil || !currentEmu.Running {
				js.Global().Get("console").Call("warn", "postMessage: emulator not running, ignoring input")
				return nil
			}

			btn := data.Get("button").String()
			pressed := false
			if p := data.Get("pressed"); !p.IsUndefined() {
				pressed = p.Bool()
			}
			js.Global().Get("console").Call("log", "emu-input payload:", btn, pressed)

			st := input.GetState()
			if st == nil {
				js.Global().Get("console").Call("warn", "postMessage: input state not initialized")
				return nil
			}

			switch btn {
			case "up":
				st.Up = pressed
			case "down":
				st.Down = pressed
			case "left":
				st.Left = pressed
			case "right":
				st.Right = pressed
			case "a":
				st.A = pressed
			case "b":
				st.B = pressed
			case "start":
				st.Start = pressed
			case "select":
				st.Select = pressed
			default:
				js.Global().Get("console").Call("warn", "message handler: unknown button", btn)
			}
		}
		return nil
	})
	js.Global().Set("emuMessageHandler", msgHandler)
	js.Global().Call("addEventListener", "message", msgHandler)

	// Expose stopEmulator function to allow JS to stop the running emulator
	stopEmulator := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if currentEmu != nil && currentEmu.Running {
			logger.Info("stopEmulator called from JS - stopping emulator")
			currentEmu.Running = false
			return true
		}
		logger.Warn("stopEmulator called but no emulator is running")
		return false
	})
	js.Global().Set("stopEmulator", stopEmulator)

	// Main loop: wait for ROMs to start. This keeps the main goroutine alive
	// and ensures UiInit (which calls ebiten.RunGame) runs on the main thread.
	for {
		romConfig := <-romStartCh
		logger.Info("Starting emulator from enqueued ROM (%d bytes)", len(romConfig.ROMBytes))
		js.Global().Get("console").Call("log", "🎮 ROM received, size:", len(romConfig.ROMBytes))

		// Set the DMG color palette type before starting the emulator
		// Default to "auto" if not specified (enables GBC colorization for DMG games)
		colorMode := romConfig.ColorMode
		if colorMode == "" {
			colorMode = "auto"
		}
		ui.SetDMGColorsPaletteType(colorMode)
		logger.Info("Set color mode to: %s", colorMode)
		js.Global().Get("console").Call("log", "🎨 Color mode set to:", colorMode)

		emuInstance := ui.StartEmulatorFromBytes(romConfig.ROMBytes)
		js.Global().Get("console").Call("log", "✅ Emulator instance created")

		// Save the current emu instance for debug reads
		currentEmu = emuInstance
		// Run the UI (blocks until the emulator stops)
		ui.UiInit(emuInstance, false)
		logger.Info("UiInit returned; emulator stopped or exited")
		currentEmu = nil
	}
}
