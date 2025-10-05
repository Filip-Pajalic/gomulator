package ui

import (
	"app/internal/cpu"
	"app/internal/input"
	"app/internal/logger"
	"errors"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 160 // Game Boy screen width
	ScreenHeight = 144 // Game Boy screen height
	scale        = 4
)

type Game struct {
	EmuCtx     *EmuContext
	VideoImage *ebiten.Image
	BootAnim   *BootAnimation
	// Debug variables
	debugImage   *ebiten.Image
	frameCounter int
}

func NewGame(emuInstance *EmuContext) *Game {
	LcdInit()

	input.LcdReadFunc = LcdRead
	input.LcdWriteFunc = LcdWrite

	// Initialize boot animation
	bootAnim := NewBootAnimation(emuInstance.BusCtx)

	g := &Game{
		EmuCtx:     emuInstance,
		VideoImage: ebiten.NewImage(ScreenWidth, ScreenHeight),
		BootAnim:   bootAnim,
	}

	//debugScale := 3 // Increased from 2 for better visibility
	//g.debugImage = ebiten.NewImage(16*8*debugScale, 24*8*debugScale) // 24 rows of tiles

	// Start boot animation and delay boot ROM simulation
	if bootAnim.Enabled {
		bootAnim.Start()
	} else {
		// If animation is disabled, run boot sequence immediately
		bootRomContext := cpu.BootRomCtx()
		bootRomContext.SimulateBootSequence()
	}

	return g
}

func (g *Game) Update() error {
	// Handle boot animation first
	if !g.BootAnim.IsComplete() {
		g.BootAnim.Update()
		g.handleBootInput() // Special input handling during boot

		// When boot animation completes, run the boot ROM sequence
		if g.BootAnim.IsComplete() && g.BootAnim.Enabled {
			bootRomContext := cpu.BootRomCtx()
			bootRomContext.SimulateBootSequence()
		}
		return nil
	}

	if !g.EmuCtx.Running {
		return ErrEmulationStopped
	}

	// Normal game operation
	g.handleInput()
	g.EmuCtx.StepFrame()

	if !g.EmuCtx.Running {
		return ErrEmulationStopped
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// If boot animation is active, draw it in the main game area
	if !g.BootAnim.IsComplete() {
		// Create a temporary image for the boot animation
		bootImage := ebiten.NewImage(ScreenWidth, ScreenHeight)
		g.BootAnim.Draw(bootImage)

		// Draw the boot animation scaled up in the main game area
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Scale(float64(scale), float64(scale))
		screen.DrawImage(bootImage, opts)

		// Still show debug windows during boot
		// g.updateDebugWindows()
		// debugOpts := &ebiten.DrawImageOptions{}
		// debugOpts.GeoM.Translate(float64(ScreenWidth*scale+10), 0)
		// screen.DrawImage(g.debugImage, debugOpts)
		return
	}

	// Normal game drawing
	screen.Fill(color.RGBA{20, 20, 20, 255})

	g.drawVideoBuffer(screen)

	// Draw debug windows on the right side
	//g.updateDebugWindows()
	debugOpts := &ebiten.DrawImageOptions{}
	debugOpts.GeoM.Translate(float64(ScreenWidth*scale+10), 0)
	//screen.DrawImage(g.debugImage, debugOpts)
}

// Layout defines the screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	debugScale := 3                   // Match the debug scale
	totalWidth := ScreenWidth * scale // Main screen + debug window + padding
	totalHeight := ScreenHeight * scale
	if debugHeight := 24 * 8 * debugScale; debugHeight > totalHeight {
		totalHeight = debugHeight
	}
	return totalWidth, totalHeight
}

func (g *Game) handleBootInput() {
	// Allow skipping boot animation with any key press
	if ebiten.IsKeyPressed(ebiten.KeySpace) ||
		ebiten.IsKeyPressed(ebiten.KeyEnter) ||
		ebiten.IsKeyPressed(ebiten.KeyEscape) {
		g.BootAnim.Skip()
	}
}

func (g *Game) handleInput() {
	state := input.GetState()

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

	// Debug: Check if video buffer has any non-zero values
	nonZeroPixels := 0
	for i := 0; i < len(videoBuffer); i++ {
		if videoBuffer[i] != 0 {
			nonZeroPixels++
		}
	}

	// Log very rarely to avoid spam
	g.frameCounter++
	if g.frameCounter%300 == 0 { // Reduced from every 60 frames to every 300 frames
		logger.Debug("Video buffer: %d non-zero pixels, VideoImage size: %dx%d",
			nonZeroPixels, g.VideoImage.Bounds().Dx(), g.VideoImage.Bounds().Dy())
	}

	// Clear the video image first
	g.VideoImage.Clear()

	// Clear to fully transparent so game pixels draw exactly as produced
	g.VideoImage.Fill(color.RGBA{0, 0, 0, 0})

	// Draw actual game content on top if available
	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			pixelValue := videoBuffer[y*ScreenWidth+x]
			if pixelValue != 0 { // Only draw non-black pixels from game
				col := convertColor(pixelValue)
				g.VideoImage.Set(x, y, col)
			}
		}
	}

	// Draw the scaled game image with a simple border
	gameOpts := &ebiten.DrawImageOptions{}
	gameOpts.GeoM.Scale(scale, scale)

	// Draw a simple border by drawing the game area twice - first larger, then smaller
	borderOpts := &ebiten.DrawImageOptions{}
	borderOpts.GeoM.Scale(scale+0.1, scale+0.1) // Slightly larger for border effect
	borderOpts.GeoM.Translate(-2, -2)

	// Create a white border image
	borderImg := ebiten.NewImage(ScreenWidth, ScreenHeight)
	borderImg.Fill(color.RGBA{255, 255, 255, 255})
	screen.DrawImage(borderImg, borderOpts)

	// Draw the actual game image on top
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
func UiInit(emuInstance *EmuContext) {
	game := NewGame(emuInstance)
	ebiten.SetWindowSize(ScreenWidth*scale, ScreenHeight*scale)
	ebiten.SetWindowTitle("Gomulator")
	if err := ebiten.RunGame(game); err != nil {
		if errors.Is(err, ErrEmulationStopped) {
			logger.Info("Emulation stopped")
			return
		}
		logger.Fatal("Failed to run game: %v", err)
	}
}
