package ui

import (
	"app/internal/gamepad"
	"app/internal/logger"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 160 // Game Boy screen width
	ScreenHeight = 144 // Game Boy screen height
	scale        = 4
)

// Game represents the game state
type Game struct {
	EmuCtx     *EmuContext
	VideoImage *ebiten.Image
	// Debug variables
	debugImage *ebiten.Image
}

// NewGame initializes a new Game instance
func NewGame(emuInstance *EmuContext) *Game {
	LcdInit()
	g := &Game{
		EmuCtx:     emuInstance,
		VideoImage: ebiten.NewImage(ScreenWidth, ScreenHeight),
	}

	// Create the debug image
	g.debugImage = ebiten.NewImage(16*8*scale, 32*8*scale)
	return g
}

// Update handles game logic and input
func (g *Game) Update() error {
	// Handle input
	g.handleInput()

	// Step the emulator for one frame
	g.EmuCtx.StepFrame()

	return nil
}

// Draw renders the game screen
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear the screen
	screen.Fill(color.Black)

	// Draw the emulator's video buffer
	g.drawVideoBuffer(screen)

	// Draw debug windows if needed
	g.updateDebugWindows()
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(ScreenWidth*scale+10), 0)
	screen.DrawImage(g.debugImage, opts)
}

// Layout defines the screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	totalWidth := ScreenWidth*scale + 16*8*scale + 10 // Main screen + debug window + padding
	totalHeight := ScreenHeight * scale
	return totalWidth, totalHeight
}

// handleInput processes user input
func (g *Game) handleInput() {
	state := gamepad.GetState()

	state.B = ebiten.IsKeyPressed(ebiten.KeyZ)
	state.A = ebiten.IsKeyPressed(ebiten.KeyX)
	state.Start = ebiten.IsKeyPressed(ebiten.KeyEnter)
	state.Select = ebiten.IsKeyPressed(ebiten.KeyTab)
	state.Up = ebiten.IsKeyPressed(ebiten.KeyArrowUp)
	state.Down = ebiten.IsKeyPressed(ebiten.KeyArrowDown)
	state.Left = ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
	state.Right = ebiten.IsKeyPressed(ebiten.KeyArrowRight)
}

// drawVideoBuffer renders the emulator's video buffer to the screen
func (g *Game) drawVideoBuffer(screen *ebiten.Image) {
	videoBuffer := g.EmuCtx.PpuCtx.VideBuffer()

	// Set pixels on the VideoImage
	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			pixelValue := videoBuffer[y*ScreenWidth+x]
			col := convertColor(pixelValue)
			g.VideoImage.Set(x, y, col)
		}
	}

	// Draw the image onto the screen with scaling
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(scale, scale)
	screen.DrawImage(g.VideoImage, opts)
}

// updateDebugWindows updates and draws the debug window
func (g *Game) updateDebugWindows() {
	// Clear the debug image
	g.debugImage.Clear()

	// Draw tiles or debug information
	var tileNum uint16 = 0
	var xDraw, yDraw int

	for y := 0; y < 24; y++ {
		for x := 0; x < 16; x++ {
			g.displayTile(g.debugImage, 0x8000, tileNum, xDraw+(x*8*scale), yDraw+(y*8*scale))
			tileNum++
		}
		yDraw += 8 * scale
		xDraw = 0
	}
}

// displayTile draws a single tile onto the given image
func (g *Game) displayTile(img *ebiten.Image, startLocation uint16, tileNum uint16, x int, y int) {
	ppu := g.EmuCtx.PpuCtx
	tileAddr := startLocation + tileNum*16
	tileImage := ebiten.NewImage(8*scale, 8*scale)
	for ty := 0; ty < 8; ty++ {
		b1 := ppu.WramRead(tileAddr + uint16(ty*2))
		b2 := ppu.WramRead(tileAddr + uint16(ty*2+1))
		for tx := 0; tx < 8; tx++ {
			bit := 7 - tx
			colorIdx := ((b2>>bit)&1)<<1 | ((b1 >> bit) & 1)
			var col color.Color
			switch colorIdx {
			case 0:
				col = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF} // White
			case 1:
				col = color.RGBA{0xAA, 0xAA, 0xAA, 0xFF} // Light gray
			case 2:
				col = color.RGBA{0x55, 0x55, 0x55, 0xFF} // Dark gray
			case 3:
				col = color.RGBA{0x00, 0x00, 0x00, 0xFF} // Black
			}
			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					tileImage.Set(tx*scale+sx, ty*scale+sy, col)
				}
			}
		}
	}
	rect := &ebiten.DrawImageOptions{}
	rect.GeoM.Translate(float64(x), float64(y))
	img.DrawImage(tileImage, rect)
}

// convertColor converts a pixel value to a color.RGBA
func convertColor(value uint32) color.RGBA {
	return color.RGBA{
		R: uint8((value >> 16) & 0xFF),
		G: uint8((value >> 8) & 0xFF),
		B: uint8(value & 0xFF),
		A: 0xFF,
	}
}

// UiInit initializes the UI and starts the game loop
func UiInit(emuInstance *EmuContext) {
	game := NewGame(emuInstance)
	ebiten.SetWindowSize(ScreenWidth*scale+16*8*scale+10, ScreenHeight*scale)
	ebiten.SetWindowTitle("Emulator")
	if err := ebiten.RunGame(game); err != nil {
		logger.Fatal("Failed to run game:", err)
	}
}
