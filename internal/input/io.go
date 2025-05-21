package input

import (
	"app/internal/common"
	"app/internal/logger"
)

var serialData [2]byte
var ly byte = 0

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
	case 0xFF01:
		return serialData[0]
	case 0xFF02:
		return serialData[1]
	case 0xFF0F:
		return i.cpu.CpuGetIntFlags()
	case 0xFF44:
		return ly
	default:
		if common.Between16(address, 0xFF04, 0xFF07) {
			return i.timer.Read(address)
		}
		logger.Warn("UNSUPPORTED bus_read(%04X)\n", address)
		return 0
	}
}

func (i *Io) Write(address uint16, value byte) {
	switch address {
	case 0xFF01:
		serialData[0] = value
	case 0xFF02:
		serialData[1] = value
	case 0xFF0F:
		i.cpu.CpuSetIntFlags(value)
	case 0xFF46:
		i.dma.RestartDMAContext(value)
		logger.Info("DMA START!\n")
	default:
		if common.Between16(address, 0xFF04, 0xFF07) {
			i.timer.Write(address, value)
		} else {
			logger.Warn("UNSUPPORTED bus_write(%04X)\n", address)
		}
	}
}
