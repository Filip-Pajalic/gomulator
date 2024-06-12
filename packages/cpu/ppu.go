package cpu

import (
	"bytes"
	"encoding/binary"
	log "pajalic.go.emulator/packages/logger"
)

func PpuInit() {

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
}

var ppuCtx = PpuContext{OamRam: oamEntry}

func PpuWramWrite(address uint16, value byte) {
	ppuCtx.Vram[address-0x8000] = value
}

func PpuWramRead(address uint16) byte {
	return ppuCtx.Vram[address-0x8000]
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

func PpuOamWrite(address uint16, value byte) {
	if address >= 0xFE00 {
		address -= 0xFE00
	}
	ppuCtx.OamRam[address] = DecodeToOamEntry(EncodeToBytes(oamEntry[address]))
}

func PpuOamRead(address uint16) byte {
	if address >= 0xFE00 {
		address -= 0xFE00
	}
	return EncodeToBytes(oamEntry[address])[address]

}

func EncodeToBytes(entry OamEntry) []byte {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, &entry)
	if err != nil {
		log.Fatal(err.Error())
	}
	return buf.Bytes()
}

func DecodeToOamEntry(bytearray []byte) OamEntry {
	reader := bytes.NewReader(bytearray)

	var entry OamEntry
	err := binary.Read(reader, binary.BigEndian, &entry)
	if err != nil {
		log.Fatal(err.Error())
	}
	return entry
}
