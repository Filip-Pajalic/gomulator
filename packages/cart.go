package gameboypackage

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"unsafe"
)

type cartContext struct {
	filename [1024]byte
	romSize  uint32
	romData  []byte
	header   *romHeader
}

type romHeader struct {
	Entry          [4]byte
	Logo           [0x30]byte
	Title          [16]byte
	NewLicCode     [2]byte
	SgbFlag        byte
	CartType       byte
	RomSize        byte
	RamSize        byte
	DestCode       byte
	LicCode        byte
	Version        byte
	Checksum       byte
	GlobalChecksum [2]byte
}

var ROM_TYPES = [...][]byte{
	[]byte("ROM ONLY"),
	[]byte("MBC1"),
	[]byte("MBC1+RAM"),
	[]byte("MBC1+RAM+BATTERY"),
	[]byte("0x04 ???"),
	[]byte("MBC2"),
	[]byte("MBC2+BATTERY"),
	[]byte("0x07 ???"),
	[]byte("ROM+RAM 1"),
	[]byte("ROM+RAM+BATTERY 1"),
	[]byte("0x0A ???"),
	[]byte("MMM01"),
	[]byte("MMM01+RAM"),
	[]byte("MMM01+RAM+BATTERY"),
	[]byte("0x0E ???"),
	[]byte("MBC3+TIMER+BATTERY"),
	[]byte("MBC3+TIMER+RAM+BATTERY 2"),
	[]byte("MBC3"),
	[]byte("MBC3+RAM 2"),
	[]byte("MBC3+RAM+BATTERY 2"),
	[]byte("0x14 ???"),
	[]byte("0x15 ???"),
	[]byte("0x16 ???"),
	[]byte("0x17 ???"),
	[]byte("0x18 ???"),
	[]byte("MBC5"),
	[]byte("MBC5+RAM"),
	[]byte("MBC5+RAM+BATTERY"),
	[]byte("MBC5+RUMBLE"),
	[]byte("MBC5+RUMBLE+RAM"),
	[]byte("MBC5+RUMBLE+RAM+BATTERY"),
	[]byte("0x1F ???"),
	[]byte("MBC6"),
	[]byte("0x21 ???"),
	[]byte("MBC7+SENSOR+RUMBLE+RAM+BATTERY"),
}

var LIC_CODE = map[int][]byte{
	0x00: []byte("None"),
	0x01: []byte("Nintendo R&D1"),
	0x08: []byte("Capcom"),
	0x13: []byte("Electronic Arts"),
	0x18: []byte("Hudson Soft"),
	0x19: []byte("b-ai"),
	0x20: []byte("kss"),
	0x22: []byte("pow"),
	0x24: []byte("PCM Complete"),
	0x25: []byte("san-x"),
	0x28: []byte("Kemco Japan"),
	0x29: []byte("seta"),
	0x30: []byte("Viacom"),
	0x31: []byte("Nintendo"),
	0x32: []byte("Bandai"),
	0x33: []byte("Ocean/Acclaim"),
	0x34: []byte("Konami"),
	0x35: []byte("Hector"),
	0x37: []byte("Taito"),
	0x38: []byte("Hudson"),
	0x39: []byte("Banpresto"),
	0x41: []byte("Ubi Soft"),
	0x42: []byte("Atlus"),
	0x44: []byte("Malibu"),
	0x46: []byte("angel"),
	0x47: []byte("Bullet-Proof"),
	0x49: []byte("irem"),
	0x50: []byte("Absolute"),
	0x51: []byte("Acclaim"),
	0x52: []byte("Activision"),
	0x53: []byte("American sammy"),
	0x54: []byte("Konami"),
	0x55: []byte("Hi tech entertainment"),
	0x56: []byte("LJN"),
	0x57: []byte("Matchbox"),
	0x58: []byte("Mattel"),
	0x59: []byte("Milton Bradley"),
	0x60: []byte("Titus"),
	0x61: []byte("Virgin"),
	0x64: []byte("LucasArts"),
	0x67: []byte("Ocean"),
	0x69: []byte("Electronic Arts"),
	0x70: []byte("Infogrames"),
	0x71: []byte("Interplay"),
	0x72: []byte("Broderbund"),
	0x73: []byte("sculptured"),
	0x75: []byte("sci"),
	0x78: []byte("THQ"),
	0x79: []byte("Accolade"),
	0x80: []byte("misawa"),
	0x83: []byte("lozc"),
	0x86: []byte("Tokuma Shoten Intermedia"),
	0x87: []byte("Tsukuda Original"),
	0x91: []byte("Chunsoft"),
	0x92: []byte("Video system"),
	0x93: []byte("Ocean/Acclaim"),
	0x95: []byte("Varie"),
	0x96: []byte("Yonezawa/s’pal"),
	0x97: []byte("Kaneko"),
	0x99: []byte("Pack in soft"),
	0xA4: []byte("Konami (Yu-Gi-Oh!)"),
}

var ctx cartContext

const headerOffset = 0x100

func cartLicName() []byte {

	if ctx.header.CartType <= 0xA4 {
		return LIC_CODE[int(ctx.header.LicCode)]
	}

	return nil
}

func cartTypeName() []byte {
	if ctx.header.CartType <= 0x22 {
		return ROM_TYPES[ctx.header.CartType]
	}
	return nil
}

func checkSumChecker(checksum byte) string {
	var x uint16 = 0
	for i := uint16(0x134); i <= 0x14C; i++ {
		x = x - uint16(ctx.romData[i]) - 1
	}
	var result string

	if byte(x&0xFF) == checksum {
		result = "PASSED"

	} else {
		result = "FAILED"
	}
	return result
}

func readNextBytes(file *os.File, number int, offset int64) []byte {
	bytes := make([]byte, number)

	_, err := file.ReadAt(bytes, offset)
	if err != nil {
		Logger.Fatal(err)
	}
	return bytes
}

func cartLoad(cart string) bool {
	copy(ctx.filename[:], fmt.Sprintf("%s", cart))

	file, err := os.Open(cart)
	if err != nil {
		Logger.Fatalf("Error while opening file", err)
	}
	defer file.Close()
	Logger.Infof("Opened: %s\n", ctx.filename)

	fi, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	ctx.romSize = uint32(fi.Size())

	var lines []byte
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		temp := scanner.Bytes()
		lines = append(lines, temp...)
	}
	ctx.romData = lines

	rh := romHeader{}

	data := readNextBytes(file, int(unsafe.Sizeof(rh)), headerOffset)

	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.LittleEndian, &rh)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
	ctx.header = &rh
	ctx.header.Title[15] = 0
	Logger.Info("Cartridge Loaded:")
	Logger.Infof("Title    : %s", string(ctx.header.Title[:]))
	Logger.Infof("Type     : %2.2X (%s)", ctx.header.CartType, cartTypeName())
	Logger.Infof("ROM Size : %d KB", 32<<ctx.header.RomSize)
	Logger.Infof("RAM Size : %2.2X", ctx.header.RamSize)
	Logger.Infof("LIC Code : %2.2X (%s)", ctx.header.LicCode, cartLicName())
	Logger.Infof("ROM Vers : %2.2X", ctx.header.Version)
	Logger.Infof(
		"Checksum : %2.2X (%s)",
		ctx.header.Checksum,
		checkSumChecker(ctx.header.Checksum),
	)
	return true
}

func CartWrite(address uint16, data byte) {

	//ctx.romData[address] = data

}

func CartRead(address uint16) byte {
	return ctx.romData[address]
}
