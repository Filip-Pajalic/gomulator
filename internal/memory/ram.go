package memory

import (
	"os"
	logger "app/internal/logger"
)

// RamContext represents the state of WRAM and HRAM
type RamContext struct {
	Wram [0x2000]byte // 8KB WRAM (0xC000 - 0xDFFF)
	Hram [0x80]byte   // 128B HRAM (0xFF80 - 0xFFFE)
}

// singleton instance of RamContext
var ramInstance *RamContext

func RamCtx() *RamContext {
	if ramInstance == nil {
		ramInstance = &RamContext{}
	}
	return ramInstance
}

// WramRead reads a byte from WRAM at the given address
func (r *RamContext) WramRead(address uint16) byte {
	if address < 0xC000 || address >= 0xE000 {
		logger.Warn("WRAM Read: Invalid address %04X", address)
		return 0xFF // Return default value for invalid addresses
	}
	return r.Wram[address-0xC000]
}

// WramWrite writes a byte to WRAM at the given address
func (r *RamContext) WramWrite(address uint16, value byte) {
	if address < 0xC000 || address >= 0xE000 {
		logger.Warn("WRAM Write: Invalid address %04X", address)
		return
	}
	r.Wram[address-0xC000] = value
}

// HramRead reads a byte from HRAM at the given address
func (r *RamContext) HramRead(address uint16) byte {
	if address < 0xFF80 || address > 0xFFFE {
		logger.Warn("HRAM Read: Invalid address %04X", address)
		os.Exit(1)
	}
	return r.Hram[address-0xFF80]
}

func (r *RamContext) HramWrite(address uint16, value byte) {
	if address < 0xFF80 || address > 0xFFFE {
		logger.Warn("HRAM Write: Invalid address %04X", address)
		os.Exit(1)
	}
	r.Hram[address-0xFF80] = value
}
