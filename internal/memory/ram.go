package memory

import (
	logger "app/internal/logger"
)

const (
	enableHramDebug = false
)

// RamContext represents the state of WRAM and HRAM
type RamContext struct {
	Wram            [0x8000]byte // GBC: 32KB WRAM = 8 banks × 4KB (0xC000 - 0xDFFF)
	Hram            [0x80]byte   // 128B HRAM (0xFF80 - 0xFFFE)
	CurrentWramBank byte         // GBC: Current WRAM bank (1-7, default 1)
}

// singleton instance of RamContext
var ramInstance *RamContext

func RamCtx() *RamContext {
	if ramInstance == nil {
		ramInstance = &RamContext{
			CurrentWramBank: 1, // GBC: Default to bank 1
		}
	}
	return ramInstance
}

// WramRead reads a byte from WRAM at the given address
func (r *RamContext) WramRead(address uint16) byte {
	if address < 0xC000 || address >= 0xE000 {
		logger.Warn("WRAM Read: Invalid address %04X", address)
		return 0xFF // Return default value for invalid addresses
	}

	// GBC: Handle WRAM banking
	if address < 0xD000 {
		// 0xC000-0xCFFF: Bank 0 (always accessible)
		return r.Wram[address-0xC000]
	} else {
		// 0xD000-0xDFFF: Switchable bank (1-7)
		offset := address - 0xD000
		bankOffset := uint16(r.CurrentWramBank) * 0x1000
		return r.Wram[bankOffset+offset]
	}
}

// WramWrite writes a byte to WRAM at the given address
func (r *RamContext) WramWrite(address uint16, value byte) {
	if address < 0xC000 || address >= 0xE000 {
		logger.Warn("WRAM Write: Invalid address %04X", address)
		return
	}
	if address == 0xD807 || address == 0xD808 {
		logger.Debug("WRAM write debug: addr=%04X value=%02X bank=%d", address, value, r.CurrentWramBank)
	}

	// GBC: Handle WRAM banking
	if address < 0xD000 {
		// 0xC000-0xCFFF: Bank 0 (always accessible)
		r.Wram[address-0xC000] = value
	} else {
		// 0xD000-0xDFFF: Switchable bank (1-7)
		offset := address - 0xD000
		bankOffset := uint16(r.CurrentWramBank) * 0x1000
		r.Wram[bankOffset+offset] = value
	}
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

// ReadWramBank returns the current WRAM bank register value (0xFF70)
func ReadWramBank() byte {
	return RamCtx().CurrentWramBank
}

// WriteWramBank sets the WRAM bank register (0xFF70)
func WriteWramBank(value byte) {
	bank := value & 0x07 // Only bits 0-2 are used
	if bank == 0 {
		bank = 1 // Bank 0 maps to bank 1
	}
	RamCtx().CurrentWramBank = bank
}
