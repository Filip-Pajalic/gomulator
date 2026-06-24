//go:build js && wasm

package ui

import (
	"app/internal/input"

	"github.com/hajimehoshi/ebiten/v2"
)

var wasmJSInput input.State

// SetWASMInput updates the browser-control input state. JS calls into this
// through cmd/wasm.go; handleInputPlatform combines it with iframe keyboard.
func SetWASMInput(button string, pressed bool) {
	switch button {
	case "up":
		wasmJSInput.Up = pressed
	case "down":
		wasmJSInput.Down = pressed
	case "left":
		wasmJSInput.Left = pressed
	case "right":
		wasmJSInput.Right = pressed
	case "a":
		wasmJSInput.A = pressed
	case "b":
		wasmJSInput.B = pressed
	case "start":
		wasmJSInput.Start = pressed
	case "select":
		wasmJSInput.Select = pressed
	}
}

// ResetWASMInput clears host-page input state between ROM loads.
func ResetWASMInput() {
	wasmJSInput = input.State{}
}

// handleInputPlatform handles platform-specific input logic for WASM
func handleInputPlatform(state *input.State) {
	// JS host-page controls and iframe keyboard are separate input sources.
	// Keep JS state separate so keyboard polling does not erase host events.
	kbA := ebiten.IsKeyPressed(ebiten.KeyZ)
	kbB := ebiten.IsKeyPressed(ebiten.KeyX)
	kbStart := ebiten.IsKeyPressed(ebiten.KeyEnter)
	kbSelect := ebiten.IsKeyPressed(ebiten.KeyTab)
	kbUp := ebiten.IsKeyPressed(ebiten.KeyArrowUp)
	kbDown := ebiten.IsKeyPressed(ebiten.KeyArrowDown)
	kbLeft := ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
	kbRight := ebiten.IsKeyPressed(ebiten.KeyArrowRight)

	state.B = wasmJSInput.B || kbB
	state.A = wasmJSInput.A || kbA
	state.Start = wasmJSInput.Start || kbStart
	state.Select = wasmJSInput.Select || kbSelect
	state.Up = wasmJSInput.Up || kbUp
	state.Down = wasmJSInput.Down || kbDown
	state.Left = wasmJSInput.Left || kbLeft
	state.Right = wasmJSInput.Right || kbRight
}
