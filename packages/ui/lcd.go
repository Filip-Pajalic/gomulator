package ui

import "pajalic.go.emulator/packages/cpu"

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

var colorsDefault = [4]uint32{0xFFFFFFFF, 0xFFAAAAAA, 0xFF555555, 0xFF000000}

// lcd_mode represents the different LCD modes.
type lcdMode uint8

const (
	ModeHBlank lcdMode = iota
	ModeVBlank
	ModeOam
	ModeXfer
)

// LcdInit initializes the LCD context.
func LcdInit() {
	lcdContext.Lcdc = 0x91
	lcdContext.Lcds = 0
	lcdContext.ScrollX = 0
	lcdContext.ScrollY = 0
	lcdContext.Ly = 0
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
}

// LcdCtx returns the LCD context.
func LcdCtx() *LcdContext {
	return &lcdContext
}

// LcdRead reads a byte from the LCD memory.
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
		// Handle color palette access or return 0 for out-of-bounds access.
		return 0
	}
}

// UpdatePalette updates the color palette.
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
}

// LcdWrite writes a byte to the LCD memory.
func LcdWrite(address uint16, value uint8) {
	offset := address - 0xFF40
	switch offset {
	case 0:
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
		cpu.RestartDMAContext(value)
	}

	switch address {
	case 0xFF47:
		UpdatePalette(value, 0)
	case 0xFF48:
		UpdatePalette(value&0b11111100, 1)
	case 0xFF49:
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
