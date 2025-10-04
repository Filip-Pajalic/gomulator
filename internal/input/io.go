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
	case 0xFF50:
		// Boot ROM disable register - always return 0x01 (boot ROM disabled after simulation)
		return 0x01
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
	case 0xFF50:
		// Boot ROM disable register - write of any value disables boot ROM
		// Since we simulate boot sequence, this is handled but not needed
		logger.Debug("Boot ROM disable register written: %02X", value)
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
