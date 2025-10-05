package ui

import (
	"app/internal/common"
	"app/internal/logger"
	"app/internal/memory"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type BootAnimationState int

const (
	BootAnimationIdle BootAnimationState = iota
	BootAnimationFalling
	BootAnimationPause
	BootAnimationComplete
)

type BootAnimation struct {
	State          BootAnimationState
	LogoY          float64
	LogoTargetY    float64
	LogoVelocity   float64
	AnimationTimer int
	LogoTiles      [][]byte // Nintendo logo tile data
	Enabled        bool
	Bus            *memory.Bus // Access to memory bus for reading ROM logo data
}

func NewBootAnimation(bus *memory.Bus) *BootAnimation {
	return &BootAnimation{
		State:          BootAnimationIdle,
		LogoY:          -40, // Start higher above screen
		LogoTargetY:    60,  // Position in upper third of screen
		LogoVelocity:   0,
		AnimationTimer: 0,
		Enabled:        common.IsBootAnimationEnabled(),
		Bus:            bus,
	}
}

func (ba *BootAnimation) Start() {
	if !ba.Enabled {
		ba.State = BootAnimationComplete
		return
	}

	logger.Info("Boot Animation: Starting Nintendo logo animation")
	ba.State = BootAnimationFalling
	ba.LogoY = -32
	ba.LogoVelocity = 0
	ba.AnimationTimer = 0
}

func (ba *BootAnimation) Update() {
	if !ba.Enabled || ba.State == BootAnimationComplete {
		return
	}

	ba.AnimationTimer++

	// Auto-complete after 300 frames (5 seconds at 60fps) - longer timeout
	if ba.AnimationTimer > 300 {
		ba.State = BootAnimationComplete
		logger.Info("Boot Animation: Auto-completed after timeout")
		return
	}

	switch ba.State {
	case BootAnimationFalling:
		// Apply gravity (slower than before)
		ba.LogoVelocity += 0.3 * common.GlobalConfig.BootAnimationSpeed
		ba.LogoY += ba.LogoVelocity * common.GlobalConfig.BootAnimationSpeed

		// Check if logo reached target position
		if ba.LogoY >= ba.LogoTargetY {
			ba.LogoY = ba.LogoTargetY
			ba.LogoVelocity = 0 // Stop completely, no bounce
			ba.State = BootAnimationPause
			ba.AnimationTimer = 0 // Reset timer for pause
			logger.Debug("Boot Animation: Logo landed, starting pause")
		}

	case BootAnimationPause:
		// Pause for 120 frames (2 seconds) to allow screenshot
		if ba.AnimationTimer > 120 {
			ba.State = BootAnimationComplete
			logger.Info("Boot Animation: Pause complete, animation finished")
		}
	}
}

func (ba *BootAnimation) Draw(screen *ebiten.Image) {
	if !ba.Enabled || ba.State == BootAnimationComplete {
		return
	}

	// Clear screen with Game Boy boot color (light cream/white)
	screen.Fill(color.RGBA{248, 248, 248, 255})

	// Draw the Nintendo logo at current position
	logoWidth := 64                        // Approximate width of the text logo
	logoX := (ScreenWidth - logoWidth) / 2 // Center the logo
	ba.drawNintendoLogo(screen, logoX, int(ba.LogoY))
}

func (ba *BootAnimation) drawNintendoLogo(screen *ebiten.Image, x, y int) {
	logoPattern := []string{
		"██    ██ ██ ██    ██ ████████ ███████ ██    ██ ████████   ███   ",
		"███   ██ ██ ███   ██    ██    ██      ███   ██ ██     ██ ██   ██",
		"████  ██ ██ ████  ██    ██    ██      ████  ██ ██     ██ ██   ██",
		"██ ██ ██ ██ ██ ██ ██    ██    █████   ██ ██ ██ ██     ██ ██   ██",
		"██  ████ ██ ██  ████    ██    ██      ██  ████ ██     ██ ██   ██",
		"██   ███ ██ ██   ███    ██    ██      ██   ███ ██     ██ ██   ██",
		"██    ██ ██ ██    ██    ██    ███████ ██    ██ ████████   ███   ",
	}

	// Draw each pixel of the logo
	for row, line := range logoPattern {
		for col, char := range line {
			if char == '█' {
				pixelX := x + col
				pixelY := y + row*2 // Double height for better visibility

				// Draw pixel in dark Game Boy green
				if pixelX >= 0 && pixelX < ScreenWidth &&
					pixelY >= 0 && pixelY < ScreenHeight {
					screen.Set(pixelX, pixelY, color.RGBA{15, 56, 15, 255})
					// Add pixel below for thickness
					if pixelY+1 < ScreenHeight {
						screen.Set(pixelX, pixelY+1, color.RGBA{15, 56, 15, 255})
					}
				}
			}
		}
	}
}

func (ba *BootAnimation) getNintendoLogoTiles() [][]byte {

	return [][]byte{
		// Row 1 of tiles
		{0xCE, 0xED, 0x66, 0x66, 0xCC, 0x0D, 0x00, 0x0B}, // N (left)
		{0x03, 0x73, 0x00, 0x83, 0x00, 0x0C, 0x00, 0x0D}, // I
		{0x00, 0x08, 0x11, 0x1F, 0x88, 0x89, 0x00, 0x0E}, // N (right)
		{0xDC, 0xCC, 0x6E, 0xE6, 0xDD, 0xDD, 0xD9, 0x99}, // T
		// Row 2 of tiles
		{0xBB, 0xBB, 0x67, 0x63, 0x6E, 0x0E, 0xEC, 0xCC}, // E
		{0xDD, 0xDC, 0x99, 0x9F, 0xBB, 0xB9, 0x33, 0x3E}, // N
		{0x3C, 0x42, 0xB9, 0xA5, 0xB9, 0xA5, 0x42, 0x3C}, // D
		{0x4F, 0x7F, 0x00, 0x0F, 0x20, 0x3F, 0x40, 0x5F}, // O
		// Row 3 of tiles (® symbol and padding)
		{0x9F, 0xBF, 0x00, 0x1F, 0x40, 0x5F, 0x80, 0x9F}, // ®
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // Empty
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // Empty
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // Empty
	}
}

func (ba *BootAnimation) drawTile(screen *ebiten.Image, x, y int, tileData []byte) {
	for row := 0; row < 8; row++ {
		if row < len(tileData) {
			tileByte := tileData[row]
			for col := 0; col < 8; col++ {
				// Extract pixel value from each bit (monochrome)
				pixelBit := (tileByte >> (7 - col)) & 1

				if pixelBit == 1 {
					pixelX := x + col*2 // Scale 2x for visibility
					pixelY := y + row*2

					// Draw 2x2 pixel block in classic Game Boy dark green
					for dx := 0; dx < 2; dx++ {
						for dy := 0; dy < 2; dy++ {
							drawX := pixelX + dx
							drawY := pixelY + dy
							if drawX >= 0 && drawX < ScreenWidth &&
								drawY >= 0 && drawY < ScreenHeight {
								screen.Set(drawX, drawY, color.RGBA{15, 56, 15, 255})
							}
						}
					}
				}
			}
		}
	}
}

func (ba *BootAnimation) IsComplete() bool {
	return !ba.Enabled || ba.State == BootAnimationComplete
}

func (ba *BootAnimation) Skip() {
	ba.State = BootAnimationComplete
	logger.Info("Boot Animation: Skipped")
}
