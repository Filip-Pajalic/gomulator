package ui

import (
	"app/internal/cpu"
	logger "app/internal/logger"
	"bytes"
	"encoding/binary"
)

type ExternalPins interface {
	RequestInterrupt(t cpu.InterruptType)
}

const (
	LINES_PER_FRAME = 154
	TICKS_PER_LINE  = 456
	YRES            = 144
	XRES            = 160
)

// PPU interface defines PPU operations
type PPU interface {
	VramWrite(address uint16, value byte)
	VramRead(address uint16) byte
	OamWrite(address uint16, value byte)
	OamRead(address uint16) byte
	VideBuffer() []uint32
	PpuTick()
}

type PpuContext struct {
	OamRam [40]OamEntry
	Vram   [8192]byte

	LineSpriteCount   uint
	Pfc               PixelFifoContext
	LineSprites       *OamLineEntry
	LineEntryArray    [10]OamLineEntry
	FetchedEntryCount byte
	FetchedEntries    [3]OamEntry
	WindowLine        byte
	CurrentFrame      uint32
	LineTicks         uint32
	VideoBuffer       []uint32
}

var ppuInstance *PpuContext

type OamLineEntry struct {
	Entry OamEntry
	Next  *OamLineEntry
}

type OamEntry struct {
	Y            byte
	X            byte
	Tile         byte
	FCgbPn       int32
	FCgbVramBank int32
	FPn          int32
	FXFlip       int32
	FYFlip       int32
	FBgp         int32
}

// PixelFifoContext represents the PPU's pixel FIFO context
type PixelFifoContext struct {
	CurFetchState FetchState
	PixelFifo     Fifo

	LineX   uint8
	PushedX uint8
	FetchX  uint8

	BgwFetchData   [3]uint8
	FetchEntryData [6]uint8
	MapX           uint8
	MapY           byte
	TileY          byte
	FifoX          byte
}

type FetchState int

const (
	FS_TILE FetchState = iota
	FS_DATA0
	FS_DATA1
	FS_IDLE
	FS_PUSH
)

// FifoEntry represents a single FIFO entry
type FifoEntry struct {
	Next       *FifoEntry
	Value      uint32 // 32-bit color value
	ColorIndex uint8  // Original color index (0-3)
}

// Fifo represents a FIFO queue
type Fifo struct {
	head *FifoEntry
	tail *FifoEntry
	size uint32
}

// NewPpuContext initializes a new PPU context
func NewPpuContext() *PpuContext {
	logger.Debug("PPU: Initializing new PPU context")
	ctx := &PpuContext{
		VideoBuffer: make([]uint32, YRES*XRES),
		Pfc: PixelFifoContext{
			CurFetchState: FS_TILE,
			PixelFifo: Fifo{
				size: 0,
			},
		},
	}

	// Clear VRAM initially - let the ROM load its own tile data
	for i := range ctx.Vram {
		ctx.Vram[i] = 0
	}

	logger.Debug("PPU: Initialized video buffer %dx%d (%d bytes)", XRES, YRES, len(ctx.VideoBuffer))
	logger.Debug("PPU: VRAM cleared, ready for ROM tile data")
	return ctx
}

func (p *PpuContext) VideBuffer() []uint32 {
	return p.VideoBuffer
}

// PpuCtx returns the singleton PPU context
func PpuCtx() *PpuContext {
	if ppuInstance == nil {
		ppuInstance = NewPpuContext()
	}
	return ppuInstance
}

// VramWrite writes a byte to VRAM
func (p *PpuContext) VramWrite(address uint16, value byte) {
	if address >= 0x8000 && address < 0xA000 {
		p.Vram[address-0x8000] = value
		// Log writes to tile data area more frequently during early frames
		if address >= 0x8000 && address < 0x9800 && p != nil && PpuCtx().CurrentFrame < 100 {
			if address%64 == 0 || value != 0 { // Log every 64th address or any non-zero write
				logger.Debug("VRAM WRITE: Tile data at %04X = %02X (frame %d)", address, value, PpuCtx().CurrentFrame)
			}
		}
	} else {
		logger.Warn("PPU VramWrite: Invalid address %04X", address)
	}
}

// VramRead reads a byte from VRAM
func (p *PpuContext) VramRead(address uint16) byte {
	if address >= 0x8000 && address < 0xA000 {
		val := p.Vram[address-0x8000]
		// Debug: log tile data access occasionally
		if address < 0x8100 && address%64 == 0 {
			logger.Debug("VRAM READ: %04X = %02X", address, val)
		}
		return val
	}
	// Reduce warning spam - only log occasionally for invalid addresses
	if address%256 == 0 {
		logger.Debug("PPU VramRead: Invalid address %04X", address)
	}
	return 0xFF
}

// OamWrite writes a byte to OAM
func (p *PpuContext) OamWrite(address uint16, value byte) {
	if address >= 0xFE00 && address < 0xFEA0 {
		offset := address - 0xFE00
		entryIndex := offset / 4 // Each OamEntry is 4 bytes
		fieldOffset := offset % 4

		if entryIndex >= uint16(len(p.OamRam)) {
			logger.Warn("PPU OamWrite: Invalid OAM entry index %d", entryIndex)
			return
		}

		entry := &p.OamRam[entryIndex]

		switch fieldOffset {
		case 0: // Y position
			entry.Y = value
		case 1: // X position
			entry.X = value
		case 2: // Tile number
			entry.Tile = value
		case 3: // Attributes byte
			entry.FBgp = int32((value >> 7) & 1)         // Bit 7: BG/Window over OBJ (0=No, 1=BG/Win above OBJ)
			entry.FYFlip = int32((value >> 6) & 1)       // Bit 6: Y flip
			entry.FXFlip = int32((value >> 5) & 1)       // Bit 5: X flip
			entry.FPn = int32((value >> 4) & 1)          // Bit 4: Palette number (0=OBP0, 1=OBP1)
			entry.FCgbVramBank = int32((value >> 3) & 1) // Bit 3: VRAM Bank (CGB only)
			entry.FCgbPn = int32(value & 0x07)           // Bits 2-0: CGB Palette number
		}
	} else {
		logger.Warn("PPU OamWrite: Invalid address %04X", address)
	}
}

// OamRead reads a byte from OAM
func (p *PpuContext) OamRead(address uint16) byte {
	if address >= 0xFE00 && address < 0xFEA0 {
		offset := address - 0xFE00
		entryIndex := offset / 4 // Each OamEntry is 4 bytes for the first three fields
		fieldOffset := offset % 4

		if entryIndex >= uint16(len(p.OamRam)) {
			logger.Warn("PPU OamRead: Invalid OAM entry index %d", entryIndex)
			return 0xFF
		}

		entryBytes := EncodeToBytes(p.OamRam[entryIndex])
		return entryBytes[fieldOffset]
	}
	logger.Warn("PPU OamRead: Invalid address %04X", address)
	return 0xFF
}

// EncodeToBytes encodes an OamEntry to a byte slice
func EncodeToBytes(entry OamEntry) []byte {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, &entry)
	if err != nil {
		logger.Fatal("PPU EncodeToBytes failed: %v", err)
	}
	return buf.Bytes()
}

// DecodeToOamEntry decodes a byte slice to an OamEntry
func DecodeToOamEntry(data []byte) OamEntry {
	var entry OamEntry
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, &entry)
	if err != nil {
		logger.Fatal("PPU DecodeToOamEntry failed: %v", err)
	}
	return entry
}

// LCDC register bit masks
const (
	LCDC_DISPLAY_ENABLE = 0x80
	LCDC_WIN_MAP        = 0x40
	LCDC_WIN_ENABLE     = 0x20
	LCDC_TILE_DATA      = 0x10
	LCDC_BG_MAP         = 0x08
	LCDC_OBJ_SIZE       = 0x04
	LCDC_OBJ_ENABLE     = 0x02
	LCDC_BG_ENABLE      = 0x01
)

// LCDCTileSelect returns the current tile map selection
func LCDCTileSelect() bool {
	return LcdCtx().Lcdc&LCDC_BG_MAP != 0
}

// LCDCTileDataSelect returns the current tile data selection
func LCDCTileDataSelect() bool {
	return LcdCtx().Lcdc&LCDC_TILE_DATA != 0
}

// PpuTick steps the PPU forward one cycle (main state machine)
func (p *PpuContext) PpuTick() {
	// FIXED: Reference implementation NEVER checks LCD enabled in ppu_tick()
	// Always process the state machine like the reference

	// Increment line ticks FIRST like reference
	p.LineTicks++

	// Debug: Log mode changes occasionally
	currentMode := LCDSMode()
	if p.LineTicks%100 == 0 && LcdCtx().Ly < 5 {
		logger.Debug("PPU: Tick - Mode=%d LY=%d LineTicks=%d", currentMode, LcdCtx().Ly, p.LineTicks)
	}

	// Execute the current LCD mode - EXACTLY like reference ppu_tick()
	switch currentMode {
	case ModeOam:
		p.ModeOAM()
	case ModeXfer:
		p.ModePixelTransfer()
	case ModeHBlank:
		p.ModeHBlank()
	case ModeVBlank:
		p.ModeVBlank()
	}
}

// ResetLCDState resets PPU state when LCD is disabled
func (p *PpuContext) ResetLCDState() {
	p.LineTicks = 0
	LcdCtx().Ly = 0
	SetLCDMode(ModeHBlank)

	for i := range p.VideoBuffer {
		p.VideoBuffer[i] = 0xFFFFFFFF // White
	}
}

func (p *PpuContext) ModeOAM() {
	if p.LineTicks >= OAM_SCAN_TICKS {
		SetLCDMode(ModeXfer)

		// Reset FIFO state for new line (like reference)
		p.Pfc.CurFetchState = FS_TILE
		p.Pfc.LineX = 0
		p.Pfc.FetchX = 0
		p.Pfc.PushedX = 0
		p.Pfc.FifoX = 0

		logger.Debug("PPU: Line %d - OAM scan complete, found %d sprites", LcdCtx().Ly, p.LineSpriteCount)
	}

	if p.LineTicks == 1 {
		p.LineSprites = nil
		p.LineSpriteCount = 0
		p.LoadLineSprites()
	}
}

// ModePixelTransfer handles pixel transfer mode (mode 3)
func (p *PpuContext) ModePixelTransfer() {
	// Debug: Log mode transfer calls occasionally
	if LcdCtx().Ly < 2 && p.LineTicks%200 == 0 {
		logger.Debug("PPU: ModePixelTransfer called - LY=%d LineTicks=%d", LcdCtx().Ly, p.LineTicks)
	}

	// Process the pipeline every tick
	p.PipelineProcess()

	// Check if we've completed the line (pushed all 160 pixels)
	if p.Pfc.PushedX >= XRES {
		p.PipelineFifoReset()
		SetLCDMode(ModeHBlank)

		if LCDSStatInt(SSHBlank) {
			logger.Debug("PPU: H-Blank STAT interrupt requested")
		}

		logger.Debug("PPU: Line %d - Pixel transfer complete, pushed %d pixels", LcdCtx().Ly, p.Pfc.PushedX)
	}
}

func (p *PpuContext) PipelineProcess() {
	// Debug: Log pipeline process calls occasionally
	if LcdCtx().Ly < 2 && p.LineTicks%200 == 0 {
		logger.Debug("PPU: PipelineProcess called - LY=%d LineTicks=%d", LcdCtx().Ly, p.LineTicks)
	}

	p.Pfc.MapY = (LcdCtx().Ly + LcdCtx().ScrollY)
	p.Pfc.MapX = (p.Pfc.FetchX + LcdCtx().ScrollX)

	// Don't set it here as it depends on what we're actually rendering

	// Only fetch on even ticks (every 2 dots)
	if (p.LineTicks & 1) == 0 {
		p.PixelFetch()
	}

	p.PipelinePushPixel()
}

func (p *PpuContext) PipelineFifoReset() {
	for p.Pfc.PixelFifo.size > 0 {
		p.PixelFifoPop()
	}
	p.Pfc.PixelFifo.head = nil
	p.Pfc.PixelFifo.tail = nil
}

func (p *PpuContext) ModeHBlank() {
	if p.LineTicks >= TICKS_PER_LINE {
		p.LineTicks = 0

		p.IncrementLY()

		if LcdCtx().Ly >= YRES {
			// Entered V-blank
			SetLCDMode(ModeVBlank)

			// Request V-Blank interrupt (like reference)
			cpu.CpuRequestInterrupt(cpu.IT_VBLANK)
			logger.Debug("PPU: V-Blank interrupt requested")

			if LCDSStatInt(SSVBlank) {
				// Request V-Blank STAT interrupt
				logger.Debug("PPU: V-Blank STAT interrupt requested")
			}

			p.CurrentFrame++
			logger.Debug("PPU: Entering V-blank at line %d, frame %d", LcdCtx().Ly, p.CurrentFrame)
		} else {
			// Start OAM scan for next line
			SetLCDMode(ModeOam)
		}
	}
}

func (p *PpuContext) ModeVBlank() {
	if p.LineTicks >= TICKS_PER_LINE {
		p.LineTicks = 0
		p.IncrementLY()

		// Check if V-blank is complete
		if LcdCtx().Ly >= LINES_PER_FRAME {
			// Frame complete, reset to line 0 and start OAM scan
			LcdCtx().Ly = 0
			p.WindowLine = 0 // Reset window line counter for new frame
			SetLCDMode(ModeOam)
			p.CurrentFrame++

			logger.Debug("PPU: Frame %d complete, starting new frame at line 0", p.CurrentFrame)
		}
	}
}

// ResetPipelineState resets the pixel FIFO state for a new line
func (p *PpuContext) ResetPipelineState() {
	p.Pfc.LineX = 0
	p.Pfc.PushedX = 0
	p.Pfc.FetchX = 0
	p.Pfc.FifoX = 0
	p.Pfc.CurFetchState = FS_TILE

	// Clear the pixel FIFO
	p.Pfc.PixelFifo.head = nil
	p.Pfc.PixelFifo.tail = nil
	p.Pfc.PixelFifo.size = 0
}

// PixelFetch implements the pixel fetch state machine
func (p *PpuContext) PixelFetch() {
	// Debug: Log fetch state occasionally
	if LcdCtx().Ly < 2 && p.LineTicks%200 == 0 {
		logger.Debug("PPU: PixelFetch state=%d FetchX=%d", p.Pfc.CurFetchState, p.Pfc.FetchX)
	}

	switch p.Pfc.CurFetchState {
	case FS_TILE:
		p.FetchTileNumber()
	case FS_DATA0:
		p.FetchTileData0()
	case FS_DATA1:
		p.FetchTileData1()
	case FS_PUSH:
		p.PushPixelsToFIFO()
	case FS_IDLE:
		// Wait state - just advance to next state
		p.Pfc.CurFetchState = FS_PUSH
	}
}

// FetchTileNumber fetches the tile number from the background map
func (p *PpuContext) FetchTileNumber() {

	windowVisible := false
	if LCDCWinEnable() && LcdCtx().Ly >= LcdCtx().WinY && LcdCtx().WinX < 167 {
		windowX := int(LcdCtx().WinX) - 7
		if windowX <= int(p.Pfc.FetchX) {
			windowVisible = true
		}
	}

	var mapAddr uint16
	var tileMapIndex int

	if windowVisible {
		// Render window tile
		windowTileY := int(p.WindowLine) / 8
		windowTileX := (int(p.Pfc.FetchX) - (int(LcdCtx().WinX) - 7)) / 8

		p.Pfc.TileY = byte((int(p.WindowLine) % 8) * 2)

		// Bounds check
		if windowTileX >= 0 && windowTileX < 32 {
			tileMapIndex = windowTileY*32 + windowTileX
			mapAddr = LCDCWinMapArea()
		} else {
			// Outside window bounds, use background
			windowVisible = false
		}
	}

	if !windowVisible {
		// Render background tile
		tileY := (int(LcdCtx().Ly) + int(LcdCtx().ScrollY)) / 8
		tileX := (int(p.Pfc.FetchX) + int(LcdCtx().ScrollX)) / 8

		p.Pfc.TileY = byte(((int(LcdCtx().Ly) + int(LcdCtx().ScrollY)) % 8) * 2)

		// Wrap around the 32x32 tile map
		tileY = tileY % 32
		tileX = tileX % 32

		tileMapIndex = tileY*32 + tileX
		mapAddr = LCDCBgMapArea()
	}

	// Fetch the tile number
	p.Pfc.BgwFetchData[0] = p.VramRead(mapAddr + uint16(tileMapIndex))

	// Debug: Log tile numbers very occasionally
	if p.Pfc.FetchX <= 24 && LcdCtx().Ly == 0 && p.LineTicks%1000 == 0 {
		logger.Debug("PPU: Fetching tile - FetchX=%d tileMapIndex=%d tileNum=0x%02X mapAddr=0x%04X",
			p.Pfc.FetchX, tileMapIndex, p.Pfc.BgwFetchData[0], mapAddr)
	}

	// CRITICAL: Handle signed tile numbers (like reference implementation)
	if LCDCBGWDataArea() == 0x8800 {
		// In 0x8800 mode, tile numbers are signed, so add 128 to convert to unsigned
		p.Pfc.BgwFetchData[0] += 128
	}

	// Move to next fetch state and advance fetch_x (like reference)
	p.Pfc.CurFetchState = FS_DATA0
	p.Pfc.FetchX += 8
}

// FetchTileData0 fetches the first byte of tile data
func (p *PpuContext) FetchTileData0() {
	tileNum := p.Pfc.BgwFetchData[0]

	// Use the same logic as reference implementation
	tileDataAddr := LCDCBGWDataArea() + uint16(tileNum)*16 + uint16(p.Pfc.TileY)

	// Fetch the first byte of tile data for this row
	p.Pfc.BgwFetchData[1] = p.VramRead(tileDataAddr)

	// Move to next fetch state
	p.Pfc.CurFetchState = FS_DATA1
}

// FetchTileData1 fetches the second byte of tile data
func (p *PpuContext) FetchTileData1() {
	tileNum := p.Pfc.BgwFetchData[0]

	// Use the same logic as reference implementation
	tileDataAddr := LCDCBGWDataArea() + uint16(tileNum)*16 + uint16(p.Pfc.TileY) + 1

	// Fetch the second byte of tile data for this row
	p.Pfc.BgwFetchData[2] = p.VramRead(tileDataAddr)

	// Move to push state
	p.Pfc.CurFetchState = FS_PUSH
}

// PushPixelsToFIFO pushes 8 pixels from the fetched tile data to the FIFO
func (p *PpuContext) PushPixelsToFIFO() {
	// Only push if FIFO has space (like reference implementation)
	if p.Pfc.PixelFifo.size > 8 {
		// FIFO is full, can't add more pixels
		return
	}

	byte1 := p.Pfc.BgwFetchData[1]
	byte2 := p.Pfc.BgwFetchData[2]

	x := int(p.Pfc.FetchX) - (8 - int(LcdCtx().ScrollX%8))

	// Extract 8 pixels from the tile data
	for i := 0; i < 8; i++ {
		bit := 7 - i
		// Match reference implementation exactly
		hi := (byte1 >> bit) & 1
		lo := ((byte2 >> bit) & 1) << 1
		colorIndex := hi | lo

		// Convert to actual color using background palette
		pixelColor := LcdCtx().BgColors[colorIndex]

		// This matches reference implementation: if (!LCDC_BGW_ENABLE) color = bg_colors[0];
		// ALSO: Force disable background for lines 8-15 to hide mohawk hair (DMG-ACID2 test)
		if !LCDCBGWEnable() {
			pixelColor = LcdCtx().BgColors[0]
			colorIndex = 0
		}

		if x >= 0 {
			p.PixelFifoPushWithIndex(uint32(pixelColor), colorIndex)
			p.Pfc.FifoX++
		}

		x++
	}

	// Return success so fetch state can advance
	p.Pfc.CurFetchState = FS_TILE
}

// PixelFifoPush adds a pixel to the pixel FIFO
func (p *PpuContext) PixelFifoPush(value uint32) {
	p.PixelFifoPushWithIndex(value, 0) // Default to color index 0
}

// PixelFifoPushWithIndex adds a pixel with color index to the pixel FIFO
func (p *PpuContext) PixelFifoPushWithIndex(value uint32, colorIndex uint8) {
	entry := &FifoEntry{
		Value:      value,
		ColorIndex: colorIndex,
		Next:       nil,
	}

	if p.Pfc.PixelFifo.tail == nil {
		p.Pfc.PixelFifo.head = entry
		p.Pfc.PixelFifo.tail = entry
	} else {
		p.Pfc.PixelFifo.tail.Next = entry
		p.Pfc.PixelFifo.tail = entry
	}

	p.Pfc.PixelFifo.size++
}

// PixelData represents a pixel with both final color and original color index
type PixelData struct {
	Color      uint32 // Final rendered color
	ColorIndex uint8  // Original color index (0-3) for priority checking
	IsBgColor0 bool   // True if this is background color 0
}

// PixelFifoPop removes and returns a pixel from the pixel FIFO
func (p *PpuContext) PixelFifoPop() PixelData {
	if p.Pfc.PixelFifo.head == nil {
		return PixelData{Color: LcdCtx().BgColors[0], ColorIndex: 0, IsBgColor0: true}
	}

	entry := p.Pfc.PixelFifo.head
	value := entry.Value
	colorIndex := entry.ColorIndex

	p.Pfc.PixelFifo.head = p.Pfc.PixelFifo.head.Next

	if p.Pfc.PixelFifo.head == nil {
		p.Pfc.PixelFifo.tail = nil
	}

	p.Pfc.PixelFifo.size--

	return PixelData{
		Color:      value,
		ColorIndex: colorIndex,
		IsBgColor0: colorIndex == 0,
	}
}

// PipelinePushPixel pushes a pixel from the FIFO to the video buffer
func (p *PpuContext) PipelinePushPixel() {
	if p.Pfc.PixelFifo.size > 8 {
		currentLine := LcdCtx().Ly

		// Only render visible scanlines
		if currentLine < YRES {
			// Get background pixel from FIFO
			bgPixel := p.PixelFifoPop()
			finalPixel := bgPixel

			// Handle scroll X - only start rendering after scroll offset
			if p.Pfc.LineX >= (LcdCtx().ScrollX % 8) {
				// Check for sprites at this position if sprites are enabled
				if LCDCObjEnable() {
					spritePixel := p.GetSpritePixel(p.Pfc.PushedX, currentLine)
					if spritePixel.Present {
						// Handle sprite-to-background priority (CRITICAL FOR DMG-ACID2)
						if spritePixel.Priority {
							// Sprite has priority, always show sprite
							finalPixel = PixelData{Color: spritePixel.Color, ColorIndex: 0, IsBgColor0: false}
						} else {
							// Sprite only shows through color 0 of background (FBgp = 1)
							if bgPixel.IsBgColor0 {
								finalPixel = PixelData{Color: spritePixel.Color, ColorIndex: 0, IsBgColor0: false}
							}
							// Otherwise keep background pixel
						}
					}
				}

				bufferIndex := uint32(currentLine)*XRES + uint32(p.Pfc.PushedX)

				// Bounds check
				if bufferIndex < uint32(len(p.VideoBuffer)) && p.Pfc.PushedX < XRES {
					p.VideoBuffer[bufferIndex] = finalPixel.Color

					// Debug: Log pixel writes very occasionally
					if p.Pfc.PushedX%80 == 0 && currentLine%60 == 0 {
						logger.Debug("PPU: Wrote pixel 0x%08X at (%d,%d) index=%d FIFO_size=%d",
							finalPixel.Color, p.Pfc.PushedX, currentLine, bufferIndex, p.Pfc.PixelFifo.size)
					}
				}

				p.Pfc.PushedX++
			}
		}

		p.Pfc.LineX++

		// Debug: Log pixel pushing occasionally to verify rendering is happening
		if p.Pfc.PushedX%80 == 0 && currentLine%32 == 0 {
			logger.Debug("PPU: Pushed pixel %d on line %d (FIFO size: %d)", p.Pfc.PushedX, currentLine, p.Pfc.PixelFifo.size)
		}
	} else {
		// Debug: Log when FIFO is too small
		if LcdCtx().Ly < 10 && p.LineTicks%100 == 0 {
			logger.Debug("PPU: FIFO too small (size=%d), not pushing pixels", p.Pfc.PixelFifo.size)
		}
	}
}

// SpritePixel represents a sprite pixel with priority information
type SpritePixel struct {
	Color    uint32
	Priority bool // true if sprite has priority over background (FBgp = 0)
	Present  bool // true if sprite pixel is present (not transparent)
}

// GetSpritePixel checks if there's a sprite pixel at the given position
func (p *PpuContext) GetSpritePixel(x uint8, y uint8) SpritePixel {
	// Check all sprites on this line
	sprite := p.LineSprites
	for sprite != nil {
		entry := sprite.Entry

		// Check if this sprite covers the current pixel
		if entry.X <= x+8 && entry.X > x {
			// Calculate pixel position within sprite
			spriteX := x + 8 - entry.X
			spriteY := y + 16 - entry.Y

			// Get sprite height
			spriteHeight := LCDCObjHeight()

			// Bounds check
			if spriteX < 8 && spriteY < spriteHeight {
				// Get sprite tile data
				tileNum := entry.Tile

				// For 8x16 sprites, mask out the lower bit
				if spriteHeight == 16 {
					tileNum &= 0xFE
				}

				// Handle vertical flip
				if entry.FYFlip != 0 {
					spriteY = spriteHeight - 1 - spriteY
				}

				// Handle horizontal flip
				if entry.FXFlip != 0 {
					spriteX = 7 - spriteX
				}

				// Calculate tile data address (sprites always use 0x8000 method)
				tileDataAddr := 0x8000 + uint16(tileNum)*16 + uint16(spriteY)*2

				// Get the two bytes that define this row of the sprite
				byte1 := p.VramRead(tileDataAddr)
				byte2 := p.VramRead(tileDataAddr + 1)

				// Extract the pixel from the sprite data
				bitPosition := 7 - spriteX
				bit0 := (byte1 >> bitPosition) & 1
				bit1 := (byte2 >> bitPosition) & 1
				colorIndex := (bit1 << 1) | bit0

				// Color index 0 is transparent for sprites
				if colorIndex != 0 {
					// Get sprite palette
					var paletteColors [4]uint32
					if entry.FPn != 0 {
						paletteColors = LcdCtx().Sp2Colors
					} else {
						paletteColors = LcdCtx().Sp1Colors
					}

					return SpritePixel{
						Color:    uint32(paletteColors[colorIndex]),
						Priority: entry.FBgp == 0, // Priority when FBgp = 0
						Present:  true,
					}
				}
			}
		}

		sprite = sprite.Next
	}

	return SpritePixel{Color: 0, Priority: false, Present: false} // No sprite pixel (transparent)
}

// RenderLine renders one complete scanline of background tiles
func (p *PpuContext) RenderLine() {
	currentY := LcdCtx().Ly

	// Only render visible lines
	if currentY >= YRES {
		return
	}

	// Get the current background scroll positions
	scrollX := LcdCtx().ScrollX
	scrollY := LcdCtx().ScrollY

	// Which tile row are we on?
	tileRow := (int(currentY) + int(scrollY)) / 8
	tileY := (int(currentY) + int(scrollY)) % 8

	// Render all tiles for this scanline
	for screenX := 0; screenX < XRES; screenX++ {
		// Calculate which tile column we're in
		tileCol := (screenX + int(scrollX)) / 8
		tileX := (screenX + int(scrollX)) % 8

		// Wrap around the 32x32 tile map
		tileCol = tileCol % 32
		tileRow = tileRow % 32

		// Get the tile number from the background map
		mapAddr := uint16(0x9800) // Background map starts at 0x9800
		if LCDCTileSelect() {
			mapAddr = 0x9C00 // Use second map if bit 3 is set
		}

		tileMapIndex := tileRow*32 + tileCol
		tileNum := p.VramRead(mapAddr + uint16(tileMapIndex))

		// Get tile data address
		var tileDataAddr uint16
		if LCDCTileDataSelect() {
			// Tile data at 0x8000-0x8FFF (unsigned tile numbers)
			tileDataAddr = 0x8000 + uint16(tileNum)*16
		} else {
			// Tile data at 0x8800-0x97FF (signed tile numbers)
			if tileNum < 128 {
				tileDataAddr = 0x9000 + uint16(tileNum)*16
			} else {
				tileDataAddr = 0x8800 + uint16(tileNum-128)*16
			}
		}

		// Get the two bytes that define this row of the tile
		byte1 := p.VramRead(tileDataAddr + uint16(tileY*2))
		byte2 := p.VramRead(tileDataAddr + uint16(tileY*2+1))

		// Extract the pixel from the tile data
		bitPosition := 7 - tileX
		bit0 := (byte1 >> bitPosition) & 1
		bit1 := (byte2 >> bitPosition) & 1
		colorIndex := (bit1 << 1) | bit0

		// Convert to actual color using the background palette
		pixelColor := LcdCtx().BgColors[colorIndex]

		// Set the pixel in the video buffer
		bufferIndex := uint32(currentY)*XRES + uint32(screenX)
		if bufferIndex < uint32(len(p.VideoBuffer)) {
			p.VideoBuffer[bufferIndex] = uint32(pixelColor)
		}
	}

	// Debug: Log occasionally to see if we're rendering
	if currentY%32 == 0 {
		logger.Debug("PPU: Rendered line %d with scroll (%d,%d)", currentY, scrollX, scrollY)
	}
}
