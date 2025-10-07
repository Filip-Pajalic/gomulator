package ui

import (
	"app/internal/cpu"
	"app/internal/input"
	"app/internal/logger"
	"errors"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 160 // Game Boy screen width
	ScreenHeight = 144 // Game Boy screen height
	scale        = 4
)

type Game struct {
	EmuCtx       *EmuContext
	VideoImage   *ebiten.Image
	pixelBuffer  []byte // Reusable buffer for WritePixels
	lastFrameTime time.Time // For frame rate limiting
	showDebugInfo bool // Toggle FPS display
	f3Pressed    bool // Track F3 key state for debouncing
	// Debug variables
	debugImage   *ebiten.Image
	frameCounter int
}

func NewGame(emuInstance *EmuContext) *Game {
	LcdInit()

	input.LcdReadFunc = LcdRead
	input.LcdWriteFunc = LcdWrite

	g := &Game{
		EmuCtx:      emuInstance,
		VideoImage:  ebiten.NewImage(ScreenWidth, ScreenHeight),
		pixelBuffer: make([]byte, ScreenWidth*ScreenHeight*4),
		showDebugInfo: false, // FPS display off by default
	}

	// Run boot ROM simulation immediately
	bootRomContext := cpu.BootRomCtx()
	bootRomContext.SimulateBootSequence()

	return g
}

func (g *Game) Update() error {
	if !g.EmuCtx.Running {
		return ErrEmulationStopped
	}

	g.handleInput()
	g.EmuCtx.StepFrame()

	if !g.EmuCtx.Running {
		return ErrEmulationStopped
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Frame rate limiting: ensure we don't draw more than 60 FPS
	// This reduces CPU usage when VSync doesn't work
	targetFrameTime := time.Second / 60
	elapsed := time.Since(g.lastFrameTime)
	if elapsed < targetFrameTime {
		time.Sleep(targetFrameTime - elapsed)
	}
	g.lastFrameTime = time.Now()

	// Normal game drawing
	screen.Fill(color.RGBA{20, 20, 20, 255})

	g.drawVideoBuffer(screen)
	
	// Display FPS in top-left corner (if enabled)
	if g.showDebugInfo {
		drawDebugInfo(screen)
	}
}

// Layout defines the screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth * scale, ScreenHeight * scale
}

func (g *Game) handleInput() {
	state := input.GetState()

	// Toggle FPS display with F3 key (debounced)
	f3Current := ebiten.IsKeyPressed(ebiten.KeyF3)
	if f3Current && !g.f3Pressed {
		g.showDebugInfo = !g.showDebugInfo
		logger.Info("FPS display: %v", g.showDebugInfo)
	}
	g.f3Pressed = f3Current

	state.B = ebiten.IsKeyPressed(ebiten.KeyZ)
	state.A = ebiten.IsKeyPressed(ebiten.KeyX)
	state.Start = ebiten.IsKeyPressed(ebiten.KeyEnter)
	state.Select = ebiten.IsKeyPressed(ebiten.KeyTab)
	state.Up = ebiten.IsKeyPressed(ebiten.KeyArrowUp)
	state.Down = ebiten.IsKeyPressed(ebiten.KeyArrowDown)
	state.Left = ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
	state.Right = ebiten.IsKeyPressed(ebiten.KeyArrowRight)
}

func (g *Game) drawVideoBuffer(screen *ebiten.Image) {
	videoBuffer := g.EmuCtx.PpuCtx.VideBuffer()

	// Use pre-allocated buffer for maximum performance
	// Video buffer contains ARGB values (0xAARRGGBB format)
	for i := 0; i < len(videoBuffer); i++ {
		argb := videoBuffer[i]
		g.pixelBuffer[i*4+0] = byte((argb >> 16) & 0xFF) // R
		g.pixelBuffer[i*4+1] = byte((argb >> 8) & 0xFF)  // G
		g.pixelBuffer[i*4+2] = byte(argb & 0xFF)         // B
		g.pixelBuffer[i*4+3] = 0xFF                      // A (always opaque)
	}
	
	g.VideoImage.WritePixels(g.pixelBuffer)

	// Draw the scaled game image
	gameOpts := &ebiten.DrawImageOptions{}
	gameOpts.GeoM.Scale(scale, scale)

	// Draw the actual game image
	screen.DrawImage(g.VideoImage, gameOpts)
}

// updateDebugWindows updates and draws the debug window
func (g *Game) updateDebugWindows() {
	// Clear the debug image with a lighter background to distinguish it
	//g.debugImage.Fill(color.RGBA{40, 40, 40, 255})

	// Check if PPU context exists
	if g.EmuCtx == nil || g.EmuCtx.PpuCtx == nil {
		logger.Debug("PPU context is nil, skipping tile drawing")
		return
	}

	// Draw tiles with better scale for debug window
	debugScale := 3 // Better visibility
	var tileNum uint16 = 0
	var xDraw, yDraw int

	for y := 0; y < 24; y++ {
		for x := 0; x < 16; x++ {
			g.displayTileWithScale(g.debugImage, 0x8000, tileNum, xDraw+(x*8*debugScale), yDraw+(y*8*debugScale), debugScale)
			tileNum++
		}
		yDraw += 8 * debugScale
		xDraw = 0
	}
}

// displayTileWithScale draws a single tile onto the given image with custom scale
func (g *Game) displayTileWithScale(img *ebiten.Image, startLocation uint16, tileNum uint16, x int, y int, tileScale int) {
	ppu := g.EmuCtx.PpuCtx
	tileAddr := startLocation + tileNum*16

	// Create a small tile image
	tileImage := ebiten.NewImage(8, 8)

	for ty := 0; ty < 8; ty++ {
		b1 := ppu.VramRead(tileAddr + uint16(ty*2))
		b2 := ppu.VramRead(tileAddr + uint16(ty*2+1))

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
			tileImage.Set(tx, ty, col)
		}
	}

	// Scale and draw the tile
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(float64(tileScale), float64(tileScale))
	opts.GeoM.Translate(float64(x), float64(y))
	img.DrawImage(tileImage, opts)
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
func UiInit(emuInstance *EmuContext, showFPS bool) {
	game := NewGame(emuInstance)
	game.showDebugInfo = showFPS // Set initial FPS display state
	
	ebiten.SetWindowSize(ScreenWidth*scale, ScreenHeight*scale)
	ebiten.SetWindowTitle("Gomulator")
	ebiten.SetTPS(60)            // Cap at 60 ticks per second (Game Boy native speed)
	ebiten.SetVsyncEnabled(true) // Enable VSync to cap FPS at monitor refresh rate
	if err := ebiten.RunGame(game); err != nil {
		if errors.Is(err, ErrEmulationStopped) {
			logger.Info("Emulation stopped")
			return
		}
		logger.Fatal("Failed to run game: %v", err)
	}
}
