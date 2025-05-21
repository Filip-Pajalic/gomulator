package ui

import (
	"bytes"
	"encoding/binary"
	logger "app/internal/logger"
)

type InterruptType byte

type ExternalPins interface {
	RequestInterrupt(t InterruptType)
}

const (
	LINES_PER_FRAME = 154
	TICKS_PER_LINE  = 456
	YRES            = 144
	XRES            = 160
)

// PPU interface defines PPU operations
type PPU interface {
	WramWrite(address uint16, value byte)
	WramRead(address uint16) byte
	OamWrite(address uint16, value byte)
	OamRead(address uint16) byte
	VideBuffer() []uint32
}

// PpuContext represents the state of the PPU
// PpuContext represents the state and data of the PPU
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

// OamLineEntry represents an entry in the OAM line
type OamLineEntry struct {
	Entry OamEntry
	Next  *OamLineEntry
}

// OamEntry represents an object attribute memory entry
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
	Next  *FifoEntry
	Value uint32 // 32-bit color value
}

// Fifo represents a FIFO queue
type Fifo struct {
	head *FifoEntry
	tail *FifoEntry
	size uint32
}

// NewPpuContext initializes a new PPU context
func NewPpuContext() *PpuContext {
	return &PpuContext{
		// Initialize fields
		VideoBuffer: make([]uint32, YRES*XRES),
		Pfc: PixelFifoContext{
			CurFetchState: FS_TILE,
		},
	}
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
func (p *PpuContext) WramWrite(address uint16, value byte) {
	if address >= 0x8000 && address < 0xA000 {
		p.Vram[address-0x8000] = value
	} else {
		logger.Warn("PPU WramWrite: Invalid address %04X", address)
	}
}

// VramRead reads a byte from VRAM
func (p *PpuContext) WramRead(address uint16) byte {
	if address >= 0x8000 && address < 0xA000 {
		return p.Vram[address-0x8000]
	}
	logger.Warn("PPU WramRead: Invalid address %04X", address)
	return 0xFF
}

// OamWrite writes a byte to OAM
func (p *PpuContext) OamWrite(address uint16, value byte) {
	if address >= 0xFE00 && address < 0xFEA0 {
		offset := address - 0xFE00
		entryIndex := offset / 4 // Each OamEntry is 4 bytes for the first three fields
		fieldOffset := offset % 4

		if entryIndex >= uint16(len(p.OamRam)) {
			logger.Warn("PPU OamWrite: Invalid OAM entry index %d", entryIndex)
			return
		}

		entryBytes := EncodeToBytes(p.OamRam[entryIndex])

		// Update the specific byte
		entryBytes[fieldOffset] = value

		// Decode the bytes back to OamEntry
		p.OamRam[entryIndex] = DecodeToOamEntry(entryBytes)
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
		logger.Fatal("PPU EncodeToBytes failed:", err.Error())
	}
	return buf.Bytes()
}

// DecodeToOamEntry decodes a byte slice to an OamEntry
func DecodeToOamEntry(data []byte) OamEntry {
	var entry OamEntry
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, &entry)
	if err != nil {
		logger.Fatal("PPU DecodeToOamEntry failed:", err.Error())
	}
	return entry
}
