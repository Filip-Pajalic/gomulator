//go:build !js || !wasm

package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func drawDebugInfo(screen *ebiten.Image) {
	fps := ebiten.ActualFPS()
	tps := ebiten.ActualTPS()
	fpsText := fmt.Sprintf("FPS: %.1f TPS: %.1f", fps, tps)
	ebitenutil.DebugPrint(screen, fpsText)
}
