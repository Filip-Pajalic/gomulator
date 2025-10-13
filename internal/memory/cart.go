// cart_component.go
package memory

import (
	"bytes"
	"encoding/binary"
	"log/slog"
	"os"
	"unsafe"

	logger "app/internal/logger"
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
	romData  []byte
	header   *romHeader

	ramData    []byte // External RAM data
	romBank    int    // Current ROM bank (1-127)
	ramBank    int    // Current RAM bank (0-3)
	ramEnabled bool   // RAM enable flag
	bankMode   int    // Banking mode (0=ROM, 1=RAM)
	cgbFlag    byte   // GBC compatibility flag (0x00=DMG, 0x80=GBC compatible, 0xC0=GBC only)
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
		cartInstance = &CartContext{
			romBank:    1,     // Start with ROM bank 1 (bank 0 maps to bank 1)
			ramBank:    0,     // Start with RAM bank 0
			ramEnabled: false, // RAM disabled by default
			bankMode:   0,     // ROM banking mode by default
		}
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

// cgbModeName returns the CGB mode name based on the CGB flag
func (c *CartContext) cgbModeName() string {
	switch c.cgbFlag {
	case 0x80:
		return "GBC Compatible"
	case 0xC0:
		return "GBC Only"
	default:
		return "DMG Only"
	}
}

// IsGBCCart returns true if this cartridge supports Game Boy Color
func (c *CartContext) IsGBCCart() bool {
	return c.cgbFlag == 0x80 || c.cgbFlag == 0xC0
}

// IsGBCOnly returns true if this cartridge requires Game Boy Color
func (c *CartContext) IsGBCOnly() bool {
	return c.cgbFlag == 0xC0
}

// GetTitle returns the game title from the ROM header
func (c *CartContext) GetTitle() []byte {
	if c.header != nil {
		// Return a copy of the title to avoid external modification
		title := make([]byte, len(c.header.Title))
		copy(title, c.header.Title[:])
		return title
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

func (c *CartContext) loadCart(romName string) {
	data, err := os.ReadFile(romName)
	slog.Info("Loading ROM file:", slog.String("filename", romName))
	if err != nil {
		logger.Fatal("Failed to load ROM file: %v", err)
	}

	copy(c.filename[:], romName)

	c.romData = data

	if len(c.romData) == 0 {
		logger.Fatal("ROM file is empty.")
	}

	// Read and parse the ROM header from c.romData
	headerSize := int(unsafe.Sizeof(romHeader{}))
	if len(c.romData) < headerOffset+headerSize {
		logger.Fatal("ROM file is too small to contain a valid header.")
	}

	headerData := c.romData[headerOffset : headerOffset+headerSize]
	buffer := bytes.NewBuffer(headerData)
	rh := romHeader{}
	readErr := binary.Read(buffer, binary.LittleEndian, &rh)
	if readErr != nil {
		logger.Fatal("binary.Read failed: %v", readErr)
	}
	c.header = &rh
	c.header.Title[15] = 0 // Null-terminate the title

	// Read CGB flag at 0x0143
	if len(c.romData) > 0x0143 {
		c.cgbFlag = c.romData[0x0143]
	} else {
		c.cgbFlag = 0x00
	}

	// Log ROM information
	logger.Info("Cartridge Loaded:")
	logger.Info("Title    : %s", string(c.header.Title[:]))
	logger.Info("Cartridge Type : %02X (%s)", c.header.CartType, c.cartTypeName())
	logger.Info("ROM Size : %d KB", 32<<c.header.RomSize)
	logger.Info("RAM Size : %02X", c.header.RamSize)
	logger.Info("LIC Code : %02X (%s)", c.header.LicCode, c.cartLicName())
	logger.Info("ROM Vers : %02X", c.header.Version)
	logger.Info("CGB Flag : %02X (%s)", c.cgbFlag, c.cgbModeName())
	logger.Info(
		"Checksum : %02X (%s)",
		c.header.Checksum,
		c.checkSumChecker(c.header.Checksum),
	)

	// Optionally, log the size of the ROM data
	logger.Info("ROM data length: %d bytes", len(c.romData))

	c.initializeRAM()
}

func (c *CartContext) initializeRAM() {
	ramSizes := map[byte]int{
		0x00: 0,      // No RAM
		0x01: 2048,   // 2KB
		0x02: 8192,   // 8KB
		0x03: 32768,  // 32KB (4 banks of 8KB each)
		0x04: 131072, // 128KB (16 banks of 8KB each)
		0x05: 65536,  // 64KB (8 banks of 8KB each)
	}

	ramSize, exists := ramSizes[c.header.RamSize]
	if !exists {
		logger.Warn("Unknown RAM size code: %02X, assuming no RAM", c.header.RamSize)
		ramSize = 0
	}

	if ramSize > 0 {
		c.ramData = make([]byte, ramSize)
		logger.Info("Initialized %d bytes of external RAM", ramSize)
	} else {
		c.ramData = nil
		logger.Info("No external RAM present")
	}
}

// ProgramLoad loads a program into memory by writing to the bus
func (c *CartContext) ProgramLoad(program [][2]uint) {
	for _, v := range program {
		address := uint16(v[0])
		data := byte(v[1])
		logger.Info("Loading Program: Writing %02X to %04X", data, address)
		BusCtx().BusWrite(address, data)
	}
}

// CartLoad loads a cartridge from a file and initializes event processing
func (c *CartContext) CartLoad(cart string) bool {
	copy(c.filename[:], cart)
	c.loadCart(cart)
	return true
}

// LoadROMFromBytes loads a ROM directly from a byte slice (for WASM/JS)
func (c *CartContext) LoadROMFromBytes(romBytes []byte) bool {
	c.romData = append([]byte(nil), romBytes...)
	if len(c.romData) == 0 {
		logger.Fatal("ROM data is empty.")
		return false
	}
	// Read and parse the ROM header from c.romData
	headerSize := int(unsafe.Sizeof(romHeader{}))
	if len(c.romData) < headerOffset+headerSize {
		logger.Fatal("ROM data is too small to contain a valid header.")
		return false
	}
	headerData := c.romData[headerOffset : headerOffset+headerSize]
	buffer := bytes.NewBuffer(headerData)
	rh := romHeader{}
	err := binary.Read(buffer, binary.LittleEndian, &rh)
	if err != nil {
		logger.Fatal("binary.Read failed: %v", err)
		return false
	}
	c.header = &rh
	c.header.Title[15] = 0 // Null-terminate the title

	// Read CGB flag at 0x0143
	if len(c.romData) > 0x0143 {
		c.cgbFlag = c.romData[0x0143]
	} else {
		c.cgbFlag = 0x00
	}

	logger.Info("Cartridge Loaded from bytes:")
	logger.Info("Title    : %s", string(c.header.Title[:]))
	logger.Info("Cartridge Type : %02X", c.header.CartType)
	logger.Info("ROM Size : %d KB", 32<<c.header.RomSize)
	logger.Info("RAM Size : %02X", c.header.RamSize)
	logger.Info("LIC Code : %02X", c.header.LicCode)
	logger.Info("ROM Vers : %02X", c.header.Version)
	logger.Info("CGB Flag : %02X (%s)", c.cgbFlag, c.cgbModeName())
	logger.Info("ROM data length: %d bytes", len(c.romData))
	c.initializeRAM()
	return true
}

func (c *CartContext) CartWrite(address uint16, data byte) {
	switch {
	case address < 0x2000:
		// RAM Enable (0x0000-0x1FFF)
		c.ramEnabled = (data & 0x0F) == 0x0A
		logger.Debug("MBC1: RAM %s", map[bool]string{true: "enabled", false: "disabled"}[c.ramEnabled])

	case address < 0x4000:
		// ROM Bank Number (0x2000-0x3FFF)
		bank := int(data & 0x1F) // 5 bits for ROM bank
		if bank == 0 {
			bank = 1 // Bank 0 maps to bank 1
		}
		c.romBank = bank
		logger.Debug("MBC1: ROM bank set to %d", c.romBank)

	case address < 0x6000:
		// RAM Bank Number or Upper ROM Bank (0x4000-0x5FFF)
		if c.bankMode == 0 {
			// ROM banking mode - upper 2 bits of ROM bank
			upperBits := int(data&0x03) << 5
			c.romBank = (c.romBank & 0x1F) | upperBits
			logger.Debug("MBC1: ROM bank upper bits set, new bank: %d", c.romBank)
		} else {
			// RAM banking mode - RAM bank number
			c.ramBank = int(data & 0x03)
			logger.Debug("MBC1: RAM bank set to %d", c.ramBank)
		}

	case address < 0x8000:
		// Banking Mode Select (0x6000-0x7FFF)
		c.bankMode = int(data & 0x01)
		logger.Debug("MBC1: Banking mode set to %d", c.bankMode)

	case address >= 0xA000 && address < 0xC000:
		// External RAM Write (0xA000-0xBFFF)
		if c.ramEnabled && len(c.ramData) > 0 {
			ramAddr := int(address-0xA000) + (c.ramBank * 0x2000)
			if ramAddr < len(c.ramData) {
				c.ramData[ramAddr] = data
				logger.Debug("MBC1: RAM write %02X to bank %d, address %04X", data, c.ramBank, address)
			}
		} else {
			logger.Debug("MBC1: RAM write ignored (RAM disabled or not present)")
		}

	default:
		logger.Warn("Cart write to invalid address %04X = %02X", address, data)
	}
}

func (c *CartContext) CartRead(address uint16) byte {
	switch {
	case address < 0x4000:
		// ROM Bank 0 (0x0000-0x3FFF) - always reads from bank 0
		if int(address) < len(c.romData) {
			return c.romData[address]
		}
		return 0xFF

	case address < 0x8000:
		// Switchable ROM Bank (0x4000-0x7FFF)
		bankOffset := c.romBank * 0x4000
		romAddr := bankOffset + int(address-0x4000)
		if romAddr < len(c.romData) {
			return c.romData[romAddr]
		}
		logger.Debug("MBC1: ROM read beyond data, bank %d, address %04X", c.romBank, address)
		return 0xFF

	case address >= 0xA000 && address < 0xC000:
		// External RAM Read (0xA000-0xBFFF)
		if c.ramEnabled && len(c.ramData) > 0 {
			ramAddr := int(address-0xA000) + (c.ramBank * 0x2000)
			if ramAddr < len(c.ramData) {
				return c.ramData[ramAddr]
			}
		}
		logger.Debug("MBC1: RAM read from disabled/invalid RAM, address %04X", address)
		return 0xFF

	default:
		logger.Warn("Cart read from invalid address %04X", address)
		return 0xFF
	}
}
