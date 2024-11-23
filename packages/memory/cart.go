// cart_component.go
package memory

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"pajalic.go.emulator/packages/pubsub"
	"unsafe"

	log "pajalic.go.emulator/packages/logger"
)

// Cartridge interface defines methods for reading and writing cartridge data
type Cartridge interface {
	CartRead(address uint16) byte
	CartWrite(address uint16, data byte)
	CartLoad(cart string) bool
}

// CartContext holds the state and data of the cartridge
type CartContext struct {
	filename [1024]byte
	romSize  uint32
	romData  []byte
	header   *romHeader
}

// romHeader represents the header structure of a Game Boy ROM
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

// ROM_TYPES maps cartridge types to their string representations
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

// LIC_CODE maps license codes to their string representations
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

var cartInstance *CartContext

// CartCtx returns the singleton CartContext
func CartCtx() *CartContext {
	if cartInstance == nil {
		cartInstance = &CartContext{}
	}
	return cartInstance
}

const headerOffset = 0x100

// cartLicName returns the license name based on the license code
func (c *CartContext) cartLicName() []byte {
	if c.header.CartType <= 0xA4 {
		return LIC_CODE[int(c.header.LicCode)]
	}
	return nil
}

// cartTypeName returns the cartridge type name based on the cartridge type
func (c *CartContext) cartTypeName() []byte {
	if c.header.CartType <= 0x22 {
		return ROM_TYPES[c.header.CartType]
	}
	return nil
}

// checkSumChecker verifies the checksum of the ROM
func (c *CartContext) checkSumChecker(checksum byte) string {
	var x uint16 = 0
	for i := uint16(0x134); i <= 0x14C; i++ {
		x = x - uint16(c.romData[i]) - 1
	}
	var result string

	if byte(x&0xFF) == checksum {
		result = "PASSED"
	} else {
		result = "FAILED"
	}
	return result
}

// readNextBytes reads a specific number of bytes from a file at a given offset
func (c *CartContext) readNextBytes(file *os.File, number int, offset int64) []byte {
	bbytes := make([]byte, number)

	_, err := file.ReadAt(bbytes, offset)
	if err != nil {
		log.Fatal(err.Error())
	}
	return bbytes
}

// loadCart loads the ROM data from a file
func (c *CartContext) loadCart(romName string) {
	fi, err := os.Open(romName)
	if err != nil {
		fmt.Println(romName, "is an invalid file. Could not open.")
		panic(err)
	}
	defer fi.Close()

	memory := make([]uint8, 0, 65536)
	buf := make([]byte, 1024)
	for {
		bytesRead, err := fi.Read(buf)
		slice := buf[0:bytesRead]
		memory = append(memory, slice...) // Expand the slice

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	emptyMemory := make([]uint8, 65536-len(memory)) // Ensure 64KB memory
	c.romData = append(memory, emptyMemory...)

	// Read and parse the ROM header
	rh := romHeader{}
	data := c.readNextBytes(fi, int(unsafe.Sizeof(rh)), headerOffset)
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.LittleEndian, &rh)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
	c.header = &rh
	c.header.Title[15] = 0 // Null-terminate the title

	log.Info("Cartridge Loaded:")
	log.Info("Title    : %s", string(c.header.Title[:]))
	log.Info("Operation     : %2.2X (%s)", c.header.CartType, c.cartTypeName())
	log.Info("ROM Size : %d KB", 32<<c.header.RomSize)
	log.Info("RAM Size : %2.2X", c.header.RamSize)
	log.Info("LIC Code : %2.2X (%s)", c.header.LicCode, c.cartLicName())
	log.Info("ROM Vers : %2.2X", c.header.Version)
	log.Info(
		"Checksum : %2.2X (%s)",
		c.header.Checksum,
		c.checkSumChecker(c.header.Checksum),
	)
}

// ProgramLoad loads a program into memory by writing to the bus
func (c *CartContext) ProgramLoad(program [][2]uint) {
	for _, v := range program {
		address := uint16(v[0])
		data := byte(v[1])
		log.Info("Loading Program: Writing %02X to %04X", data, address)
		pubsub.BusCtx().BusWrite(address, data)
	}
}

// CartLoad loads a cartridge from a file and initializes event processing
func (c *CartContext) CartLoad(cart string) bool {
	copy(c.filename[:], fmt.Sprintf("%s", cart))
	c.loadCart(cart)

	// Initialize event processing after loading the cart
	c.StartCartComponent()

	return true
}

// CartWrite writes a byte to the ROM data
func (c *CartContext) CartWrite(address uint16, data byte) {
	if address >= uint16(len(c.romData)) {
		log.Warn("Attempted to write outside ROM memory at address %04X", address)
		return
	}
	c.romData[address] = data
}

// CartRead reads a byte from the ROM data
func (c *CartContext) CartRead(address uint16) byte {
	if address >= uint16(len(c.romData)) {
		log.Warn("Attempted to read outside ROM memory at address %04X", address)
		return 0xFF
	}
	return c.romData[address]
}

// StartCartComponent initializes and starts event processing for the cartridge
func (c *CartContext) StartCartComponent() {
	// Create a ReadWriteConfig for memory read/write
	config := pubsub.NewReadWriteConfig[uint16, byte](
		pubsub.MemoryReadEvent,
		pubsub.MemoryWriteEvent,
		c.CartRead,
		c.CartWrite,
	)

	// Start processing read and write events in a separate goroutine
	go pubsub.ProcessChannelTransactions(config)
}
