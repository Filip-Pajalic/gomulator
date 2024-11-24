package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"pajalic.go.emulator/packages/logger"
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

	// Update emulator state
	if !g.EmuCtx.CpuCtx.Step() {
		g.EmuCtx.Die = true
		logger.Fatal("CPU has stopped unexpectedly.")
	}

	// You might need to call ExecuteCycles here if needed
	// g.EmuCtx.ExecuteCycles(cycles)

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
	state := GamePadGetState()

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
	// Implement tile drawing logic based on your emulator's memory and tile data
	// For demonstration purposes, we'll draw a simple rectangle

	// Replace this with actual tile rendering logic
	rect := &ebiten.DrawImageOptions{}
	rect.GeoM.Translate(float64(x), float64(y))
	tileImage := ebiten.NewImage(8*scale, 8*scale)
	tileImage.Fill(color.RGBA{0xAA, 0xAA, 0xAA, 0xFF})
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
