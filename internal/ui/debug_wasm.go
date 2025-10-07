//go:build js && wasm

package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// In WASM, DebugPrint is very expensive (creates textures for text rendering)
// So we disable it to improve performance
func drawDebugInfo(screen *ebiten.Image) {
	// No-op in WASM for better performance
}
