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
	BgColors   [4]uint32 // DMG palette colors
	Sp1Colors  [4]uint32 // DMG sprite palette 1
	Sp2Colors  [4]uint32 // DMG sprite palette 2

	// GBC Color Palettes
	BgcpIndex      uint8        // Background Color Palette Index (BCPS/BGPI)
	BgcpAutoInc    bool         // Auto-increment flag for BCPS
	BgColorPalette [64]byte     // 8 palettes × 4 colors × 2 bytes (RGB555)
	ObcpIndex      uint8        // Object Color Palette Index (OCPS/OBPI)
	ObcpAutoInc    bool         // Auto-increment flag for OCPS
	ObColorPalette [64]byte     // 8 palettes × 4 colors × 2 bytes (RGB555)
	BgColorCache   [8][4]uint32 // Cached RGB888 colors for BG palettes
	ObColorCache   [8][4]uint32 // Cached RGB888 colors for sprite palettes
}

var lcdContext LcdContext

// Default colors (matching reference implementation)
var colorsDefault = [4]uint32{0xFFFFFFFF, 0xFFAAAAAA, 0xFF555555, 0xFF000000}

// GBC Default Color Palettes for DMG games
// These are the palettes the Game Boy Color uses when playing original Game Boy games
// The GBC assigns palettes based on the game's ROM header (title/manufacturer)
var (
	// Grayscale palette (neutral default, used for unknown games)
	gbcPaletteGrayscale = [4]uint32{
		0xFFFFFFFF, // White
		0xFFAAAAAA, // Light Gray
		0xFF555555, // Dark Gray
		0xFF000000, // Black
	}

	// Classic DMG Green palette (original Game Boy look)
	gbcPaletteGreen = [4]uint32{
		0xFF9BBC0F, // Lightest
		0xFF8BAC0F, // Light
		0xFF306230, // Dark
		0xFF0F380F, // Darkest
	}

	// Brown/Sepia palette (warm tones)
	gbcPaletteBrown = [4]uint32{
		0xFFFFFFCC, // Lightest - Pale Yellow
		0xFFCCA562, // Light - Tan
		0xFF8B6239, // Dark - Brown
		0xFF000000, // Darkest - Black
	}

	// Red/Pink palette (vibrant tones)
	gbcPaletteRed = [4]uint32{
		0xFFF8E8C8, // Lightest - Cream
		0xFFD89048, // Light - Orange
		0xFFA82820, // Dark - Red
		0xFF301850, // Darkest - Dark Purple
	}

	// Blue palette (cool tones)
	gbcPaletteBlue = [4]uint32{
		0xFFFFFFFF, // Lightest - White
		0xFFADD8E6, // Light - Light Blue
		0xFF4169E1, // Dark - Royal Blue
		0xFF000033, // Darkest - Dark Navy
	}
)

// GBC color mode for DMG games
var dmgGbcColorsEnabled bool
var dmgGbcPalette [4]uint32
var dmgColorsPaletteType string   // Stores the palette type choice for initialization
var dmgColorsUsedAutoLookup bool  // True if auto mode with palette lookup was used
var dmgAutoLookupPalette PaletteEntry // Stores the auto-detected palette configuration

// PaletteEntry represents a DMG palette configuration used by GBC
type PaletteEntry struct {
	BGPalette   [4]uint32
	OBJ0Palette [4]uint32
	OBJ1Palette [4]uint32
}

// CalculateTitleHash calculates the GBC boot ROM hash for a game title
// This is used to look up the appropriate color palette for DMG games
func CalculateTitleHash(title []byte) uint8 {
	// The hash is calculated from bytes 0x134-0x143 (title)
	hash := uint8(0)
	for i := 0; i < len(title) && i < 16; i++ {
		hash += title[i]
	}
	return hash
}

// GBC Palette Lookup Table
// This maps game title hashes to their assigned color palettes
// Based on the GBC boot ROM palette assignment system
// NOTE: Hashes are calculated from the ROM title (bytes 0x134-0x143)
var gbcPaletteLookup = map[uint8]PaletteEntry{
	// Zelda: Link's Awakening (title="ZELDA", hash=0x70)
	0x70: {
		BGPalette:   [4]uint32{0xFFFFFFFF, 0xFF7BFF31, 0xFF008400, 0xFF000000}, // Green theme
		OBJ0Palette: [4]uint32{0xFFFFFFFF, 0xFFFF8484, 0xFF943A3A, 0xFF000000},
		OBJ1Palette: [4]uint32{0xFFFFFFFF, 0xFF63A5FF, 0xFF0000FF, 0xFF000000},
	},
	// Tetris (hash varies by region)
	0x52: {
		BGPalette:   [4]uint32{0xFFFFFFFF, 0xFFFF9C00, 0xFFFF0000, 0xFF000000}, // White, Orange, Red, Black
		OBJ0Palette: [4]uint32{0xFFFFFFFF, 0xFFFF9C00, 0xFFFF0000, 0xFF000000},
		OBJ1Palette: [4]uint32{0xFFFFFFFF, 0xFFFF9C00, 0xFFFF0000, 0xFF000000},
	},
	// Super Mario Land
	0x14: {
		BGPalette:   [4]uint32{0xFFFFFFFF, 0xFFFFAD63, 0xFF843100, 0xFF000000}, // White, Tan, Brown, Black
		OBJ0Palette: [4]uint32{0xFFFFFFFF, 0xFFFF8484, 0xFF943A3A, 0xFF000000}, // Mario red
		OBJ1Palette: [4]uint32{0xFFFFFFFF, 0xFF7BFF31, 0xFF0063C5, 0xFF000000}, // Green/Blue
	},
	// Kirby's Dream Land
	0x27: {
		BGPalette:   [4]uint32{0xFFFFFFFF, 0xFFFFCE9C, 0xFFCE6563, 0xFF000000},
		OBJ0Palette: [4]uint32{0xFFFFFFFF, 0xFFFFAD63, 0xFFFF6B6B, 0xFF000000}, // Pink for Kirby
		OBJ1Palette: [4]uint32{0xFFFFFFFF, 0xFFFFAD63, 0xFF943100, 0xFF000000},
	},
	// Pokemon Red/Blue/Yellow (GBC-enhanced, but including for reference)
	0x61: {
		BGPalette:   [4]uint32{0xFFFFFFFF, 0xFFADD8E6, 0xFF4169E1, 0xFF000033}, // Blue tones
		OBJ0Palette: [4]uint32{0xFFFFFFFF, 0xFFFFAAAA, 0xFFFF5555, 0xFF000000}, // Red tones
		OBJ1Palette: [4]uint32{0xFFFFFFFF, 0xFFADD8E6, 0xFF4169E1, 0xFF000033},
	},
}

// Default palette for unknown games
var gbcDefaultPalette = PaletteEntry{
	BGPalette:   gbcPaletteGrayscale,
	OBJ0Palette: gbcPaletteGrayscale,
	OBJ1Palette: gbcPaletteGrayscale,
}

// LookupDMGPalette looks up the appropriate GBC color palette for a DMG game
// based on its title hash (mimics GBC boot ROM behavior)
func LookupDMGPalette(title []byte) PaletteEntry {
	hash := CalculateTitleHash(title)
	logger.Info("LCD: Game title hash = 0x%02X", hash)

	if palette, found := gbcPaletteLookup[hash]; found {
		logger.Info("LCD: Found palette configuration for hash 0x%02X", hash)
		return palette
	}

	logger.Info("LCD: No specific palette found for hash 0x%02X, using default", hash)
	return gbcDefaultPalette
}

// ApplyDMGPaletteEntry applies a full palette configuration (BG + OBJ0 + OBJ1)
func ApplyDMGPaletteEntry(entry PaletteEntry) {
	// Apply background palette
	lcdContext.BgColors = entry.BGPalette
	// Apply sprite palettes
	lcdContext.Sp1Colors = entry.OBJ0Palette
	lcdContext.Sp2Colors = entry.OBJ1Palette

	logger.Info("LCD: Applied DMG palette configuration")
	logger.Debug("  BG:   [%08X, %08X, %08X, %08X]",
		entry.BGPalette[0], entry.BGPalette[1], entry.BGPalette[2], entry.BGPalette[3])
	logger.Debug("  OBJ0: [%08X, %08X, %08X, %08X]",
		entry.OBJ0Palette[0], entry.OBJ0Palette[1], entry.OBJ0Palette[2], entry.OBJ0Palette[3])
	logger.Debug("  OBJ1: [%08X, %08X, %08X, %08X]",
		entry.OBJ1Palette[0], entry.OBJ1Palette[1], entry.OBJ1Palette[2], entry.OBJ1Palette[3])
}

// EnableDMGGBCColors enables GBC color palettes for DMG games
// paletteType: "auto", "green", "grayscale", "brown", "red", "blue" (defaults to "auto")
// title: game title bytes for auto palette detection (can be nil for manual palette selection)
func EnableDMGGBCColors(paletteType string, title []byte) {
	dmgGbcColorsEnabled = true

	switch paletteType {
	case "green":
		dmgGbcPalette = gbcPaletteGreen
		logger.Info("LCD: DMG GBC colors enabled with Green palette")
	case "grayscale", "gray":
		dmgGbcPalette = gbcPaletteGrayscale
		logger.Info("LCD: DMG GBC colors enabled with Grayscale palette")
	case "brown":
		dmgGbcPalette = gbcPaletteBrown
		logger.Info("LCD: DMG GBC colors enabled with Brown palette")
	case "red":
		dmgGbcPalette = gbcPaletteRed
		logger.Info("LCD: DMG GBC colors enabled with Red palette")
	case "blue":
		dmgGbcPalette = gbcPaletteBlue
		logger.Info("LCD: DMG GBC colors enabled with Blue palette")
	case "auto":
		fallthrough
	default:
		// Auto mode: detect palette based on game title
		if title != nil {
			paletteEntry := LookupDMGPalette(title)
			ApplyDMGPaletteEntry(paletteEntry)
			dmgColorsUsedAutoLookup = true     // Mark that we used auto lookup
			dmgAutoLookupPalette = paletteEntry // Store the palette for UpdatePalette to use
			logger.Info("LCD: DMG GBC colors enabled with Auto palette detection")
		} else {
			dmgGbcPalette = gbcPaletteGrayscale
			logger.Info("LCD: DMG GBC colors enabled with Auto palette (grayscale fallback, no title provided)")
		}
	}
}

// IsDMGGBCColorsEnabled returns whether GBC colors are enabled for DMG games
func IsDMGGBCColorsEnabled() bool {
	return dmgGbcColorsEnabled
}

// SetDMGColorsPaletteType stores the palette type choice for later initialization
func SetDMGColorsPaletteType(paletteType string) {
	dmgColorsPaletteType = paletteType
}

// GetDMGColorsPaletteType returns the stored palette type choice
func GetDMGColorsPaletteType() string {
	return dmgColorsPaletteType
}

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

	// Initialize colors based on mode
	if dmgColorsUsedAutoLookup {
		// Auto mode with lookup - colors already set via ApplyDMGPaletteEntry
		// But we still need to call UpdatePalette to ensure proper palette mapping
		logger.Debug("LCD: Colors already set by auto palette lookup, applying palette mapping")
	} else {
		// Either not using DMG colors, or using manual palette selection
		var initialColors [4]uint32
		if IsDMGGBCColorsEnabled() {
			initialColors = dmgGbcPalette
		} else {
			initialColors = colorsDefault
		}

		for i := 0; i < 4; i++ {
			lcdContext.BgColors[i] = initialColors[i]
			lcdContext.Sp1Colors[i] = initialColors[i]
			lcdContext.Sp2Colors[i] = initialColors[i]
		}

		logger.Debug("LCD: Initialized colors from defaults")
	}

	// Always update palettes to ensure proper mapping (works in all modes now)
	UpdatePalette(0xFC, 0) // Background palette
	UpdatePalette(0xFF, 1) // Sprite palette 1
	UpdatePalette(0xFF, 2) // Sprite palette 2

	SetLCDMode(ModeVBlank)

	logger.Debug("LCD: Initialized with LCDC=0x%02X, LY=0x%02X, starting in V-blank mode", lcdContext.Lcdc, lcdContext.Ly)
}

// ReadBGCPIndex reads the BCPS register
func ReadBGCPIndex() byte {
	index := lcdContext.BgcpIndex
	if lcdContext.BgcpAutoInc {
		index |= 0x80
	}
	return index
}

// WriteBGCPIndex writes to the BCPS register
func WriteBGCPIndex(value byte) {
	lcdContext.BgcpIndex = value & 0x3F
	lcdContext.BgcpAutoInc = (value & 0x80) != 0
}

// ReadOBCPIndex reads the OCPS register
func ReadOBCPIndex() byte {
	index := lcdContext.ObcpIndex
	if lcdContext.ObcpAutoInc {
		index |= 0x80
	}
	return index
}

// WriteOBCPIndex writes to the OCPS register
func WriteOBCPIndex(value byte) {
	lcdContext.ObcpIndex = value & 0x3F
	lcdContext.ObcpAutoInc = (value & 0x80) != 0
}

func LcdCtx() *LcdContext {
	return &lcdContext
}

func LcdRead(address uint16) uint8 {
	offset := address - 0xFF40
	switch offset {
	case 0:
		return lcdContext.Lcdc
		if ppuInstance != nil {
			ppuInstance.WindowLine = 0
		}
		// Add cases for other fields as needed.
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

	// Choose color source based on mode
	var colorSource [4]uint32
	if dmgColorsUsedAutoLookup {
		// Auto mode: use the stored palette entry
		switch pal {
		case 1:
			colorSource = dmgAutoLookupPalette.OBJ0Palette
		case 2:
			colorSource = dmgAutoLookupPalette.OBJ1Palette
		default:
			colorSource = dmgAutoLookupPalette.BGPalette
		}
	} else if IsDMGGBCColorsEnabled() && !IsGBCMode() {
		// Manual palette mode: use the selected palette
		colorSource = dmgGbcPalette
	} else {
		// Default grayscale
		colorSource = colorsDefault
	}

	pColors[0] = colorSource[paletteData&0b11]
	pColors[1] = colorSource[(paletteData>>2)&0b11]
	pColors[2] = colorSource[(paletteData>>4)&0b11]
	pColors[3] = colorSource[(paletteData>>6)&0b11]

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

// GBC Color Palette Functions

// RGB555ToRGB888 converts RGB555 (5-5-5 bits) to RGB888 (8-8-8 bits)
func RGB555ToRGB888(rgb555 uint16) uint32 {
	r := uint32(rgb555 & 0x1F)
	g := uint32((rgb555 >> 5) & 0x1F)
	b := uint32((rgb555 >> 10) & 0x1F)

	// Convert 5-bit to 8-bit by shifting left 3 and copying top 3 bits to bottom
	r8 := (r << 3) | (r >> 2)
	g8 := (g << 3) | (g >> 2)
	b8 := (b << 3) | (b >> 2)

	return 0xFF000000 | (r8 << 16) | (g8 << 8) | b8
}

// UpdateBgColorCache updates the cached BG color for a specific palette and color index
func UpdateBgColorCache(paletteIndex, colorIndex int) {
	// Each color is 2 bytes (RGB555)
	offset := paletteIndex*8 + colorIndex*2
	low := lcdContext.BgColorPalette[offset]
	high := lcdContext.BgColorPalette[offset+1]
	rgb555 := uint16(low) | (uint16(high) << 8)
	lcdContext.BgColorCache[paletteIndex][colorIndex] = RGB555ToRGB888(rgb555)
}

// UpdateObColorCache updates the cached sprite color for a specific palette and color index
func UpdateObColorCache(paletteIndex, colorIndex int) {
	// Each color is 2 bytes (RGB555)
	offset := paletteIndex*8 + colorIndex*2
	low := lcdContext.ObColorPalette[offset]
	high := lcdContext.ObColorPalette[offset+1]
	rgb555 := uint16(low) | (uint16(high) << 8)
	lcdContext.ObColorCache[paletteIndex][colorIndex] = RGB555ToRGB888(rgb555)
}

// ReadBGCP reads from the Background Color Palette Data register (BCPD/BGPD)
func ReadBGCP() uint8 {
	index := lcdContext.BgcpIndex & 0x3F
	return lcdContext.BgColorPalette[index]
}

// WriteBGCP writes to the Background Color Palette Data register (BCPD/BGPD)
func WriteBGCP(value uint8) {
	index := lcdContext.BgcpIndex & 0x3F
	lcdContext.BgColorPalette[index] = value

	// Update color cache
	paletteIndex := int(index / 8)
	colorIndex := int((index % 8) / 2)
	UpdateBgColorCache(paletteIndex, colorIndex)

	// Auto-increment if enabled
	if lcdContext.BgcpAutoInc {
		lcdContext.BgcpIndex = ((lcdContext.BgcpIndex + 1) & 0x3F) | (lcdContext.BgcpIndex & 0x80)
	}

	logger.Debug("GBC: BGCP[%02X] = %02X (pal=%d, col=%d) RGB888=0x%08X", index, value, paletteIndex, colorIndex, lcdContext.BgColorCache[paletteIndex][colorIndex])
}

// ReadOBCP reads from the Object Color Palette Data register (OCPD/OBPD)
func ReadOBCP() uint8 {
	index := lcdContext.ObcpIndex & 0x3F
	return lcdContext.ObColorPalette[index]
}

// WriteOBCP writes to the Object Color Palette Data register (OCPD/OBPD)
func WriteOBCP(value uint8) {
	index := lcdContext.ObcpIndex & 0x3F
	lcdContext.ObColorPalette[index] = value

	// Update color cache
	paletteIndex := int(index / 8)
	colorIndex := int((index % 8) / 2)
	UpdateObColorCache(paletteIndex, colorIndex)

	// Auto-increment if enabled
	if lcdContext.ObcpAutoInc {
		lcdContext.ObcpIndex = ((lcdContext.ObcpIndex + 1) & 0x3F) | (lcdContext.ObcpIndex & 0x80)
	}

	logger.Debug("GBC: OBCP[%02X] = %02X (pal=%d, col=%d)", index, value, paletteIndex, colorIndex)
}
