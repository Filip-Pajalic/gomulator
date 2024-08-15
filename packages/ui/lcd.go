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

var LcdCtx LcdContext

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
	LcdCtx.Lcdc = 0x91
	LcdCtx.Lcds = 0
	LcdCtx.ScrollX = 0
	LcdCtx.ScrollY = 0
	LcdCtx.Ly = 0
	LcdCtx.LyCompare = 0
	LcdCtx.Dma = 0
	LcdCtx.BgPalette = 0xFC
	LcdCtx.ObjPalette[0] = 0xFF
	LcdCtx.ObjPalette[1] = 0xFF
	LcdCtx.WinY = 0
	LcdCtx.WinX = 0

	for i := 0; i < 4; i++ {
		LcdCtx.BgColors[i] = colorsDefault[i]
		LcdCtx.Sp1Colors[i] = colorsDefault[i]
		LcdCtx.Sp2Colors[i] = colorsDefault[i]
	}
}

// LcdGetContext returns the LCD context.
func LcdGetContext() *LcdContext {
	return &LcdCtx
}

// LcdRead reads a byte from the LCD memory.
func LcdRead(address uint16) uint8 {
	offset := address - 0xFF40
	switch offset {
	case 0:
		return LcdCtx.Lcdc
	case 1:
		return LcdCtx.Lcds
	case 2:
		return LcdCtx.ScrollY
	case 3:
		return LcdCtx.ScrollX
	case 4:
		return LcdCtx.Ly
	case 5:
		return LcdCtx.LyCompare
	case 6:
		return LcdCtx.Dma
	case 7:
		return LcdCtx.BgPalette
	case 8:
		return LcdCtx.ObjPalette[0]
	case 9:
		return LcdCtx.ObjPalette[1]
	case 10:
		return LcdCtx.WinY
	case 11:
		return LcdCtx.WinX
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
		pColors = &LcdCtx.Sp1Colors
	case 2:
		pColors = &LcdCtx.Sp2Colors
	default:
		pColors = &LcdCtx.BgColors
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
		LcdCtx.Lcdc = value
	case 1:
		LcdCtx.Lcds = value
	case 2:
		LcdCtx.ScrollY = value
	case 3:
		LcdCtx.ScrollX = value
	case 4:
		LcdCtx.Ly = value
	case 5:
		LcdCtx.LyCompare = value
	case 6:
		LcdCtx.Dma = value
	case 7:
		LcdCtx.BgPalette = value
	case 8:
		LcdCtx.ObjPalette[0] = value
	case 9:
		LcdCtx.ObjPalette[1] = value
	case 10:
		LcdCtx.WinY = value
	case 11:
		LcdCtx.WinX = value
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
	return bit(LcdGetContext().Lcdc, 0)
}

func LCDCObjEnable() bool {
	return bit(LcdGetContext().Lcdc, 1)
}

func LCDCObjHeight() uint8 {
	if bit(LcdGetContext().Lcdc, 2) {
		return 16
	}
	return 8
}

func LCDCBgMapArea() uint16 {
	if bit(LcdGetContext().Lcdc, 3) {
		return 0x9C00
	}
	return 0x9800
}

func LCDCBGWDataArea() uint16 {
	if bit(LcdGetContext().Lcdc, 4) {
		return 0x8000
	}
	return 0x8800
}

func LCDCWinEnable() bool {
	return bit(LcdGetContext().Lcdc, 5)
}

func LCDCWinMapArea() uint16 {
	if bit(LcdGetContext().Lcdc, 6) {
		return 0x9C00
	}
	return 0x9800
}

func LCDCLCDEnable() bool {
	return bit(LcdGetContext().Lcdc, 7)
}

func LCDSMode() lcdMode {
	return lcdMode(LcdGetContext().Lcds & 0b11)
}

func LCDSModeSet(mode lcdMode) {
	LcdGetContext().Lcds &^= 0b11
	LcdGetContext().Lcds |= uint8(mode)
}

func LCDSLyc() bool {
	return bit(LcdGetContext().Lcds, 2)
}

func LCDSLycSet(b bool) {
	bitSet(&LcdGetContext().Lcds, 2, b)
}

type statSrc uint8

const (
	SSHBlank statSrc = 1 << 3
	SSVBlank statSrc = 1 << 4
	SSOam    statSrc = 1 << 5
	SSLyc    statSrc = 1 << 6
)

func LCDSStatInt(src statSrc) bool {
	return LcdGetContext().Lcds&uint8(src) != 0
}
