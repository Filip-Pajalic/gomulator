//go:build js && wasm

package main

import (
	"app/internal/cpu"
	"app/internal/input"
	"app/internal/logger"
	"app/internal/memory"
	"app/internal/ui"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"strings"
	"syscall/js"
)

// currentEmu holds the active emulator instance (if any) so debug JS functions
// can access the bus for ad-hoc reads.
var currentEmu *ui.EmuContext

// ROMStartConfig holds configuration for starting a ROM
type ROMStartConfig struct {
	ROMBytes  []byte
	ColorMode string // "auto", "green", "grayscale", "brown", "red", "blue", or "" for default
	SaveKey   string
}

type wasmLocalStorageSaveStore struct{}

func (wasmLocalStorageSaveStore) Load(key string) ([]byte, error) {
	storage := js.Global().Get("localStorage")
	if storage.IsUndefined() || storage.IsNull() {
		return nil, nil
	}

	value := storage.Call("getItem", key)
	if value.IsUndefined() || value.IsNull() {
		return nil, nil
	}

	return base64.StdEncoding.DecodeString(value.String())
}

func (wasmLocalStorageSaveStore) Save(key string, data []byte) error {
	storage := js.Global().Get("localStorage")
	if storage.IsUndefined() || storage.IsNull() {
		return nil
	}

	storage.Call("setItem", key, base64.StdEncoding.EncodeToString(data))
	return nil
}

func wasmSaveKey(fileName string, romBytes []byte) string {
	return fmt.Sprintf(
		"gomulator-save:%s:%08x",
		sanitizeSaveName(fileName),
		crc32.ChecksumIEEE(romBytes),
	)
}

func sanitizeSaveName(fileName string) string {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return "rom"
	}

	var b strings.Builder
	for _, r := range fileName {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
		if b.Len() >= 120 {
			break
		}
	}

	if b.Len() == 0 {
		return "rom"
	}
	return b.String()
}

func platformInit() {
	// WASM-specific initialization
	logger.Info("Running in WASM/browser mode")

	// Also log directly to JS console to verify logging works
	js.Global().Get("console").Call("log", "🔧 Go platformInit called - WASM is running!")
}

func applyWASMInput(btn string, pressed bool) bool {
	switch btn {
	case "up", "down", "left", "right", "a", "b", "start", "select":
		ui.SetWASMInput(btn, pressed)
		return true
	default:
		return false
	}
}

func wasmDebugState() js.Value {
	state := js.Global().Get("Object").New()
	state.Set("running", false)
	state.Set("joyp", fmt.Sprintf("0x%02X", input.GetOutput()))

	pressed := input.GetState()
	state.Set("pressedStart", pressed.Start)
	state.Set("pressedSelect", pressed.Select)
	state.Set("pressedA", pressed.A)
	state.Set("pressedB", pressed.B)
	state.Set("pressedUp", pressed.Up)
	state.Set("pressedDown", pressed.Down)
	state.Set("pressedLeft", pressed.Left)
	state.Set("pressedRight", pressed.Right)

	if currentEmu == nil {
		return state
	}

	state.Set("running", currentEmu.Running)
	state.Set("ticks", fmt.Sprintf("%d", currentEmu.Ticks))
	if currentEmu.BusCtx != nil {
		state.Set("ie", fmt.Sprintf("0x%02X", currentEmu.BusCtx.BusRead(0xFFFF)))
	}

	cpuCtx, ok := currentEmu.CpuCtx.(*cpu.CpuContext)
	if !ok || cpuCtx == nil {
		return state
	}

	state.Set("pc", fmt.Sprintf("0x%04X", cpuCtx.Regs.Pc))
	state.Set("sp", fmt.Sprintf("0x%04X", cpuCtx.Regs.Sp))
	state.Set("if", fmt.Sprintf("0x%02X", cpuCtx.IntFlags))
	state.Set("ime", cpuCtx.IntMasterEnabled)
	state.Set("halted", cpuCtx.Halted)
	state.Set("stopped", cpuCtx.Stopped)
	state.Set("a", fmt.Sprintf("0x%02X", cpuCtx.Regs.A))
	state.Set("f", fmt.Sprintf("0x%02X", cpuCtx.Regs.F))
	state.Set("bc", fmt.Sprintf("0x%02X%02X", cpuCtx.Regs.B, cpuCtx.Regs.C))
	state.Set("de", fmt.Sprintf("0x%02X%02X", cpuCtx.Regs.D, cpuCtx.Regs.E))
	state.Set("hl", fmt.Sprintf("0x%02X%02X", cpuCtx.Regs.H, cpuCtx.Regs.L))
	return state
}

func platformMain() {
	logger.Info("Waiting for ROM from JavaScript...")
	memory.SetSaveStore(wasmLocalStorageSaveStore{})

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

		fileName := "rom"
		if len(args) >= 3 && !args[2].IsUndefined() && !args[2].IsNull() {
			fileName = args[2].String()
		}
		saveKey := wasmSaveKey(fileName, romBytes)

		logger.Info("ROM received from JS (%d bytes), enqueuing for start...", len(romBytes))
		// Enqueue the ROM configuration for the main goroutine to pick up and start the UI
		select {
		case romStartCh <- ROMStartConfig{ROMBytes: romBytes, ColorMode: colorMode, SaveKey: saveKey}:
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

		if !applyWASMInput(btn, pressed) {
			js.Global().Get("console").Call("warn", "emuInput: unknown button", btn)
		}

		return nil
	})
	// Keep reference in global so it won't be garbage collected
	js.Global().Set("emuInput", emuInput)

	debugState := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		state := wasmDebugState()
		js.Global().Get("console").Call("table", state)
		return state
	})
	js.Global().Set("gomulatorDebugState", debugState)

	flushSave := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		memory.FlushSave()
		return true
	})
	js.Global().Set("flushSave", flushSave)

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
			btn := data.Get("button").String()
			pressed := false
			if p := data.Get("pressed"); !p.IsUndefined() {
				pressed = p.Bool()
			}

			if !applyWASMInput(btn, pressed) {
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
			memory.FlushSave()
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
		ui.ResetWASMInput()

		// Set the DMG color palette type before starting the emulator
		// Default to "auto" if not specified (enables GBC colorization for DMG games)
		colorMode := romConfig.ColorMode
		if colorMode == "" {
			colorMode = "auto"
		}
		ui.SetDMGColorsPaletteType(colorMode)
		logger.Info("Set color mode to: %s", colorMode)
		js.Global().Get("console").Call("log", "🎨 Color mode set to:", colorMode)

		emuInstance := ui.StartEmulatorFromBytesWithSaveKey(romConfig.ROMBytes, romConfig.SaveKey)
		js.Global().Get("console").Call("log", "✅ Emulator instance created")

		// Save the current emu instance for debug reads
		currentEmu = emuInstance
		// Run the UI (blocks until the emulator stops)
		ui.UiInit(emuInstance, false)
		memory.FlushSave()
		ui.ResetWASMInput()
		logger.Info("UiInit returned; emulator stopped or exited")
		currentEmu = nil
	}
}
