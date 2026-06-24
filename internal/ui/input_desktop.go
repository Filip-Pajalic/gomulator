//go:build !js || !wasm

package ui

import (
	"app/internal/input"

	"github.com/hajimehoshi/ebiten/v2"
)

// handleInputPlatform handles platform-specific input logic for desktop
func handleInputPlatform(state *input.State) {
	// On desktop, keyboard is the only input source - directly assign state
	state.B = ebiten.IsKeyPressed(ebiten.KeyZ)
	state.A = ebiten.IsKeyPressed(ebiten.KeyX)
	state.Start = ebiten.IsKeyPressed(ebiten.KeyEnter)
	state.Select = ebiten.IsKeyPressed(ebiten.KeyTab)
	state.Up = ebiten.IsKeyPressed(ebiten.KeyArrowUp)
	state.Down = ebiten.IsKeyPressed(ebiten.KeyArrowDown)
	state.Left = ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
	state.Right = ebiten.IsKeyPressed(ebiten.KeyArrowRight)
}
