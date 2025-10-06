package ui

import (
	"app/internal/cpu"
	logger "app/internal/logger"
)

// LcdContext represents the LCD context with all its registers and color palettes.
type LcdContext struct {
	Lcdc       uint8
	Lcds       uint8
	ScrollY    uint8
	ScrollX    uint8
	Ly         uint8
	LyCompare  uint8
	Dma        uint8
	BgPalette  uint8
	ObjPalette [2]uint8
	WinY       uint8
	WinX       uint8
	BgColors   [4]uint32
	Sp1Colors  [4]uint32
	Sp2Colors  [4]uint32
}

var lcdContext LcdContext

// Default colors (matching reference implementation)
var colorsDefault = [4]uint32{0xFFFFFFFF, 0xFFAAAAAA, 0xFF555555, 0xFF000000}

// lcd_mode represents the different LCD modes.
type lcdMode uint8

const (
	ModeHBlank lcdMode = 0 // H-blank
	ModeVBlank lcdMode = 1 // V-blank
	ModeOam    lcdMode = 2 // OAM scan
	ModeXfer   lcdMode = 3 // Pixel transfer
)

// PPU timing constants - using specific names to avoid conflicts
const (
	OAM_SCAN_TICKS   = 80
	PIXEL_XFER_TICKS = 172
	HBLANK_TICKS     = 204
)

func LcdInit() {

	lcdContext.Lcdc = 0x91
	lcdContext.Lcds = 0x85 // Mode 1 (V-blank) + coincidence flag
	lcdContext.ScrollX = 0
	lcdContext.ScrollY = 0
	lcdContext.Ly = 0x91 // Boot ROM sets LY to 145 (V-blank)
	lcdContext.LyCompare = 0
	lcdContext.Dma = 0
	lcdContext.BgPalette = 0xFC
	lcdContext.ObjPalette[0] = 0xFF
	lcdContext.ObjPalette[1] = 0xFF
	lcdContext.WinY = 0
	lcdContext.WinX = 0

	for i := 0; i < 4; i++ {
		lcdContext.BgColors[i] = colorsDefault[i]
		lcdContext.Sp1Colors[i] = colorsDefault[i]
		lcdContext.Sp2Colors[i] = colorsDefault[i]
	}

	UpdatePalette(0xFC, 0) // Background palette
	UpdatePalette(0xFF, 1) // Sprite palette 1
	UpdatePalette(0xFF, 2) // Sprite palette 2

	SetLCDMode(ModeVBlank)

	logger.Debug("LCD: Initialized with LCDC=0x%02X, LY=0x%02X, starting in V-blank mode", lcdContext.Lcdc, lcdContext.Ly)
}

func LcdCtx() *LcdContext {
	return &lcdContext
}

func LcdRead(address uint16) uint8 {
	offset := address - 0xFF40
	switch offset {
	case 0:
		return lcdContext.Lcdc
	case 1:
		return lcdContext.Lcds
	case 2:
		return lcdContext.ScrollY
	case 3:
		return lcdContext.ScrollX
	case 4:
		return lcdContext.Ly
	case 5:
		return lcdContext.LyCompare
	case 6:
		return lcdContext.Dma
	case 7:
		return lcdContext.BgPalette
	case 8:
		return lcdContext.ObjPalette[0]
	case 9:
		return lcdContext.ObjPalette[1]
	case 10:
		return lcdContext.WinY
	case 11:
		return lcdContext.WinX
	// Add cases for other fields as needed.
	default:
		return 0
	}
}

func UpdatePalette(paletteData uint8, pal uint8) {
	var pColors *[4]uint32
	switch pal {
	case 1:
		pColors = &lcdContext.Sp1Colors
	case 2:
		pColors = &lcdContext.Sp2Colors
	default:
		pColors = &lcdContext.BgColors
	}

	pColors[0] = colorsDefault[paletteData&0b11]
	pColors[1] = colorsDefault[(paletteData>>2)&0b11]
	pColors[2] = colorsDefault[(paletteData>>4)&0b11]
	pColors[3] = colorsDefault[(paletteData>>6)&0b11]

	// Debug: Log palette update
	logger.Debug("LCD: Updated palette %d with data 0x%02X: [%08X, %08X, %08X, %08X]",
		pal, paletteData, pColors[0], pColors[1], pColors[2], pColors[3])
}

func LcdWrite(address uint16, value uint8) {
	offset := address - 0xFF40
	switch offset {
	case 0:
		// Debug: Log LCDC writes
		if address == 0xFF40 {
			// Special alert when background gets enabled
			if (value&0x01) != 0 && (lcdContext.Lcdc&0x01) == 0 {
				logger.Debug("*** BACKGROUND ENABLED! *** LCDC: 0x%02X -> 0x%02X", lcdContext.Lcdc, value)
			}

			// only when window is actually disabled via LCDC bit 5
			if (value&0x20) == 0 && (lcdContext.Lcdc&0x20) != 0 {
				// Window disabled via LCDC bit 5 - reset window line counter
				ppuInstance.WindowLine = 0
				logger.Debug("Window disabled via LCDC bit 5 - reset window line counter")
			}

			// Only log major changes to reduce spam
			if (value & 0x81) != (lcdContext.Lcdc & 0x81) {
				logger.Debug("LCD: LCDC write 0x%02X (LCD_EN=%v, BG_EN=%v, OBJ_EN=%v)",
					value, (value&0x80) != 0, (value&0x01) != 0, (value&0x02) != 0)
			}
		}
		lcdContext.Lcdc = value
	case 1:
		lcdContext.Lcds = value
	case 2:
		lcdContext.ScrollY = value
	case 3:
		lcdContext.ScrollX = value
	case 4:
		lcdContext.Ly = value
	case 5:
		lcdContext.LyCompare = value
		// CRITICAL FIX: Check for immediate LY=LYC match when LYC is written
		// DMG-ACID2 sets LYC=8 and expects interrupt to fire when LY=8
		if lcdContext.Ly == lcdContext.LyCompare && LCDCLCDEnabled() {
			// Trigger LY=LYC match immediately if conditions are met
			// This ensures DMG-ACID2 LYC=8 interrupt fires at the right time
			LCDSLycSet(true)
			if LCDSStatInt(SSLyc) {
				cpu.CpuRequestInterrupt(cpu.IT_LCD_STAT)
			}
		}
	case 6:
		lcdContext.Dma = value
	case 7:
		lcdContext.BgPalette = value
	case 8:
		lcdContext.ObjPalette[0] = value
	case 9:
		lcdContext.ObjPalette[1] = value
	case 10:
		lcdContext.WinY = value
	case 11:
		lcdContext.WinX = value
		// Add cases for other fields as needed.
	}

	if offset == 6 {
		// 0xFF46 = DMA
		//TODO
		//cpu.RestartDMAContext(value)
	}

	switch address {
	case 0xFF47:
		logger.Debug("LCD: Updating background palette to 0x%02X", value)
		UpdatePalette(value, 0)
	case 0xFF48:
		logger.Debug("LCD: Updating sprite palette 1 to 0x%02X", value)
		UpdatePalette(value&0b11111100, 1)
	case 0xFF49:
		logger.Debug("LCD: Updating sprite palette 2 to 0x%02X", value)
		UpdatePalette(value&0b11111100, 2)
	}
}

// Utility functions to mimic the C macros
func bit(value uint8, bit uint8) bool {
	return value&(1<<bit) != 0
}

func bitSet(value *uint8, bit uint8, set bool) {
	if set {
		*value |= (1 << bit)
	} else {
		*value &^= (1 << bit)
	}
}

func LCDCBGWEnable() bool {
	return bit(LcdCtx().Lcdc, 0)
}

func LCDCObjEnable() bool {
	return bit(LcdCtx().Lcdc, 1)
}

func LCDCObjHeight() uint8 {
	if bit(LcdCtx().Lcdc, 2) {
		return 16
	}
	return 8
}

func LCDCBgMapArea() uint16 {
	if bit(LcdCtx().Lcdc, 3) {
		return 0x9C00
	}
	return 0x9800
}

func LCDCBGWDataArea() uint16 {
	if bit(LcdCtx().Lcdc, 4) {
		return 0x8000
	}
	return 0x8800
}

func LCDCWinEnable() bool {
	return bit(LcdCtx().Lcdc, 5)
}

func LCDCWinMapArea() uint16 {
	if bit(LcdCtx().Lcdc, 6) {
		return 0x9C00
	}
	return 0x9800
}

func LCDCLCDEnable() bool {
	return bit(LcdCtx().Lcdc, 7)
}

func LCDSMode() lcdMode {
	return lcdMode(LcdCtx().Lcds & 0b11)
}

func LCDSModeSet(mode lcdMode) {
	LcdCtx().Lcds &^= 0b11
	LcdCtx().Lcds |= uint8(mode)
}

func LCDSLyc() bool {
	return bit(LcdCtx().Lcds, 2)
}

func LCDSLycSet(b bool) {
	bitSet(&LcdCtx().Lcds, 2, b)
}

type statSrc uint8

const (
	SSHBlank statSrc = 1 << 3
	SSVBlank statSrc = 1 << 4
	SSOam    statSrc = 1 << 5
	SSLyc    statSrc = 1 << 6
)

func LCDSStatInt(src statSrc) bool {
	return LcdCtx().Lcds&uint8(src) != 0
}

// LCDCLCDEnabled returns whether the LCD is enabled
func LCDCLCDEnabled() bool {
	return LCDCLCDEnable()
}

// SetLCDMode sets the LCD mode and triggers appropriate interrupts
func SetLCDMode(mode lcdMode) {
	LCDSModeSet(mode)

	// Check for STAT interrupts based on mode
	switch mode {
	case ModeHBlank:
		if LCDSStatInt(SSHBlank) {
			// TODO: Request STAT interrupt when interrupt system is ready
			logger.Debug("LCD: H-blank STAT interrupt requested")
		}
	case ModeVBlank:
		if LCDSStatInt(SSVBlank) {
			// TODO: Request STAT interrupt when interrupt system is ready
			logger.Debug("LCD: V-blank STAT interrupt requested")
		}
		// TODO: Request VBlank interrupt when interrupt system is ready
		logger.Debug("LCD: V-blank interrupt requested")
	case ModeOam:
		if LCDSStatInt(SSOam) {
			// TODO: Request STAT interrupt when interrupt system is ready
			logger.Debug("LCD: OAM STAT interrupt requested")
		}
	}
}
