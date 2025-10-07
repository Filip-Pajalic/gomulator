package memory

import (
	logger "app/internal/logger"
)

const (
	enableHramDebug = false
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
	if address == 0xD807 || address == 0xD808 {
		logger.Debug("WRAM write debug: addr=%04X value=%02X", address, value)
	}
	r.Wram[address-0xC000] = value
}

// HramRead reads a byte from HRAM at the given address
func (r *RamContext) HramRead(address uint16) byte {
	if address < 0xFF80 || address > 0xFFFE {
		logger.Warn("HRAM Read: Invalid address %04X", address)
		return 0xFF // Return 0xFF instead of exiting for better error recovery
	}
	return r.Hram[address-0xFF80]
}

func (r *RamContext) HramWrite(address uint16, value byte) {
	if address < 0xFF80 || address > 0xFFFE {
		logger.Warn("HRAM Write: Invalid address %04X", address)
		return // Just return instead of exiting for better error recovery
	}
	if enableHramDebug && address >= 0xFF80 && address <= 0xFF83 {
		logger.Debug("HRAM write debug: addr=%04X value=%02X", address, value)
	}
	r.Hram[address-0xFF80] = value
}
