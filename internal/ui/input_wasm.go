//go:build js && wasm

package ui

import (
	"app/internal/input"

	"github.com/hajimehoshi/ebiten/v2"
)

// handleInputPlatform handles platform-specific input logic for WASM
func handleInputPlatform(state *input.State) {
	// On WASM, JS handlers (via postMessage or emuInput() in cmd/wasm.go) write directly
	// to the shared input state. Since those are event-based and persist between frames,
	// we need to be careful. However, the JS handlers send BOTH press AND release events
	// (pressed=true and pressed=false), so the state should be current.

	// The issue: keyboard input also needs to work. We OR keyboard with current state
	// (which may have JS input) to allow both sources to work together.

	// Read keyboard input
	kbB := ebiten.IsKeyPressed(ebiten.KeyZ)
	kbA := ebiten.IsKeyPressed(ebiten.KeyX)
	kbStart := ebiten.IsKeyPressed(ebiten.KeyEnter)
	kbSelect := ebiten.IsKeyPressed(ebiten.KeyTab)
	kbUp := ebiten.IsKeyPressed(ebiten.KeyArrowUp)
	kbDown := ebiten.IsKeyPressed(ebiten.KeyArrowDown)
	kbLeft := ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
	kbRight := ebiten.IsKeyPressed(ebiten.KeyArrowRight)

	// Use keyboard directly (like desktop) since JS handlers update state directly
	// If we OR, keyboard releases won't work (buttons get stuck)
	state.B = kbB
	state.A = kbA
	state.Start = kbStart
	state.Select = kbSelect
	state.Up = kbUp
	state.Down = kbDown
	state.Left = kbLeft
	state.Right = kbRight
}
