package input

import (
	"app/internal/common"
	"app/internal/cpu"
	"app/internal/logger"
)

var serialData [2]byte

// Function pointers for LCD operations to avoid import cycles
var LcdReadFunc func(address uint16) uint8
var LcdWriteFunc func(address uint16, value uint8)

// GBC-specific register access functions (to be set by other packages)
var VramBankReadFunc func() byte
var VramBankWriteFunc func(byte)
var WramBankReadFunc func() byte
var WramBankWriteFunc func(byte)
var BgcpIndexReadFunc func() byte
var BgcpIndexWriteFunc func(byte)
var BgcpDataReadFunc func() byte
var BgcpDataWriteFunc func(byte)
var ObcpIndexReadFunc func() byte
var ObcpIndexWriteFunc func(byte)
var ObcpDataReadFunc func() byte
var ObcpDataWriteFunc func(byte)

type Timer interface {
	Write(address uint16, value byte)
	Read(address uint16) byte
}

type DMA interface {
	RestartDMAContext(start byte)
}

type Cpu interface {
	CpuGetIntFlags() byte
	CpuSetIntFlags(value byte)
}

type Io struct {
	cpu   Cpu
	timer Timer
	dma   DMA
}

var ioInstance *Io

func NewIo(cpu Cpu, timer Timer, dma DMA) *Io {
	ioInstance = &Io{
		cpu:   cpu,
		timer: timer,
		dma:   dma,
	}
	return ioInstance
}

func IoCtx() *Io {
	return ioInstance
}

func (i *Io) Read(address uint16) byte {
	switch address {
	case 0xFF00:
		return GetOutput()
	case 0xFF01:
		return serialData[0]
	case 0xFF02:
		return serialData[1]
	case 0xFF0F:
		if i.cpu != nil {
			return i.cpu.CpuGetIntFlags()
		} else {
			// Use global CPU function if interface is nil
			return cpu.CpuGetIntFlags()
		}
	case 0xFF44:
		// LY register should be read from LCD context, not local variable
		if LcdReadFunc != nil {
			return LcdReadFunc(address)
		}
		logger.Warn("LCD not initialized for LY read at 0xFF44")
		return 0
	case 0xFF4D:
		// GBC: KEY1 - Speed switch (not implemented yet, return normal speed)
		return 0x00
	case 0xFF4F:
		// GBC: VBK - VRAM bank select
		if VramBankReadFunc != nil {
			return VramBankReadFunc()
		}
		return 0
	case 0xFF50:
		// Boot ROM disable register - always return 0x01 (boot ROM disabled after simulation)
		return 0x01
	case 0xFF68:
		// GBC: BCPS/BGPI - Background color palette index
		if BgcpIndexReadFunc != nil {
			return BgcpIndexReadFunc()
		}
		return 0
	case 0xFF69:
		// GBC: BCPD/BGPD - Background color palette data
		if BgcpDataReadFunc != nil {
			return BgcpDataReadFunc()
		}
		return 0
	case 0xFF6A:
		// GBC: OCPS/OBPI - Object color palette index
		if ObcpIndexReadFunc != nil {
			return ObcpIndexReadFunc()
		}
		return 0
	case 0xFF6B:
		// GBC: OCPD/OBPD - Object color palette data
		if ObcpDataReadFunc != nil {
			return ObcpDataReadFunc()
		}
		return 0
	case 0xFF70:
		// GBC: SVBK - WRAM bank select
		if WramBankReadFunc != nil {
			return WramBankReadFunc()
		}
		return 1 // Default to bank 1
	default:
		if common.Between16(address, 0xFF04, 0xFF07) {
			return i.timer.Read(address)
		}
		if common.Between16(address, 0xFF40, 0xFF4B) {
			if LcdReadFunc != nil {
				return LcdReadFunc(address)
			}
			return 0
		}
		// Silently return 0 for unsupported addresses to reduce log spam
		return 0
	}
}

func (i *Io) Write(address uint16, value byte) {
	switch address {
	case 0xFF00:
		// CRITICAL FIX: Handle joypad register writes for button/direction selection
		SetSel(value)
		logger.Debug("Joypad register write: 0x%02X", value)
	case 0xFF01:
		serialData[0] = value
	case 0xFF02:
		serialData[1] = value
	case 0xFF0F:
		if i.cpu != nil {
			i.cpu.CpuSetIntFlags(value)
		} else {
			// Use global CPU function if interface is nil
			cpu.CpuSetIntFlags(value)
		}
	case 0xFF46:
		i.dma.RestartDMAContext(value)
		logger.Debug("DMA START!\n")
	case 0xFF4D:
		// GBC: KEY1 - Speed switch (not implemented yet)
		logger.Debug("GBC: Speed switch register written: %02X", value)
	case 0xFF4F:
		// GBC: VBK - VRAM bank select
		if VramBankWriteFunc != nil {
			VramBankWriteFunc(value & 0x01)
			logger.Debug("GBC: VRAM bank set to %d", value&0x01)
		}
	case 0xFF50:
		// Boot ROM disable register - write of any value disables boot ROM
		// Since we simulate boot sequence, this is handled but not needed
		logger.Debug("Boot ROM disable register written: %02X", value)
	case 0xFF68:
		// GBC: BCPS/BGPI - Background color palette index
		if BgcpIndexWriteFunc != nil {
			BgcpIndexWriteFunc(value)
			logger.Debug("GBC: BCPS written: %02X", value)
		}
	case 0xFF69:
		// GBC: BCPD/BGPD - Background color palette data
		if BgcpDataWriteFunc != nil {
			BgcpDataWriteFunc(value)
		}
	case 0xFF6A:
		// GBC: OCPS/OBPI - Object color palette index
		if ObcpIndexWriteFunc != nil {
			ObcpIndexWriteFunc(value)
			logger.Debug("GBC: OCPS written: %02X", value)
		}
	case 0xFF6B:
		// GBC: OCPD/OBPD - Object color palette data
		if ObcpDataWriteFunc != nil {
			ObcpDataWriteFunc(value)
		}
	case 0xFF70:
		// GBC: SVBK - WRAM bank select
		if WramBankWriteFunc != nil {
			bank := value & 0x07
			if bank == 0 {
				bank = 1 // Bank 0 maps to bank 1
			}
			WramBankWriteFunc(bank)
			logger.Debug("GBC: WRAM bank set to %d", bank)
		}
	default:
		if common.Between16(address, 0xFF04, 0xFF07) {
			i.timer.Write(address, value)
		} else if common.Between16(address, 0xFF40, 0xFF4B) {
			if LcdWriteFunc != nil {
				LcdWriteFunc(address, value)
			}
			// Silently ignore if LCD not initialized to reduce log spam
		} else {
			// Silently ignore unsupported writes to reduce log spam
		}
	}
}
