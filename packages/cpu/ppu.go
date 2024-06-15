package cpu

import (
	"bytes"
	"encoding/binary"
	log "pajalic.go.emulator/packages/logger"
)

const (
	LINES_PER_FRAME = 154
	TICKS_PER_LINE  = 456
	YRES            = 144
	XRES            = 160
)

func PpuInit() {
	PpuCtx.CurrentFrame = 0
	PpuCtx.LineTicks = 0

	PpuCtx.Pfc.LineX = 0
	PpuCtx.Pfc.PushedX = 0
	PpuCtx.Pfc.FetchX = 0
	PpuCtx.Pfc.PixelFifo.size = 0
	PpuCtx.Pfc.PixelFifo.head = nil
	PpuCtx.Pfc.PixelFifo.tail = nil
	PpuCtx.Pfc.CurFetchState = FS_TILE

	LcdInit()
	LCDSModeSet(ModeOam)

	for i := range PpuCtx.OamRam {
		PpuCtx.OamRam[i] = OamEntry{} // This initializes each OamEntry struct with zero values
	}

	PpuCtx.VideoBuffer = make([]uint32, YRES*XRES)

}

func PpuTick() {

}

var oamEntry [40]OamEntry

//{f_cgb_pn: 3, f_cgb_vram_bank: 1, f_pn: 1, f_y_flip: 1,f_x_flip: 1,f_bgp: 1}

/*
 Bit7   BG and Window over OBJ (0=No, 1=BG and Window colors 1-3 over the OBJ)
 Bit6   Y flip          (0=Normal, 1=Vertically mirrored)
 Bit5   X flip          (0=Normal, 1=Horizontally mirrored)
 Bit4   Palette number  **Non CGB Mode Only** (0=OBP0, 1=OBP1)
 Bit3   Tile VRAM-Bank  **CGB Mode Only**     (0=Bank 0, 1=Bank 1)
 Bit2-0 Palette number  **CGB Mode Only**     (OBP0-7)
*/

type PpuContext struct {
	OamRam [40]OamEntry
	Vram   [0x2000]byte

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

var PpuCtx = PpuContext{OamRam: oamEntry}

func PpuWramWrite(address uint16, value byte) {
	PpuCtx.Vram[address-0x8000] = value
}

func PpuWramRead(address uint16) byte {
	return PpuCtx.Vram[address-0x8000]
}

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

type FifoEntry struct {
	Next  *FifoEntry
	Value uint32 // 32-bit color value
}
type Fifo struct {
	head *FifoEntry
	tail *FifoEntry
	size uint32
}

func PpuOamWrite(address uint16, value byte) {
	if address >= 0xFE00 {
		address -= 0xFE00
	}

	entryIndex := address / 4  // Each OamEntry is 4 bytes for the initial 3 fields
	fieldOffset := address % 4 // Offset within the OamEntry

	// Encode the OamEntry to bytes
	entryBytes := EncodeToBytes(PpuCtx.OamRam[entryIndex])

	// Update the specific byte
	entryBytes[fieldOffset] = value

	// Decode the bytes back to OamEntry
	PpuCtx.OamRam[entryIndex] = DecodeToOamEntry(entryBytes)
}

func EncodeToBytes(entry OamEntry) []byte {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, &entry)
	if err != nil {
		log.Fatal(err.Error())
	}
	return buf.Bytes()
}

func DecodeToOamEntry(data []byte) OamEntry {
	var entry OamEntry
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, &entry)
	if err != nil {
		log.Fatal(err.Error())
	}
	return entry
}
func PpuOamRead(address uint16) byte {
	if address >= 0xFE00 {
		address -= 0xFE00
	}
	return EncodeToBytes(oamEntry[address])[address]

}
