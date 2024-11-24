package cpu

import (
	logger "pajalic.go.emulator/packages/logger"
)

var serialData [2]byte
var ly byte = 0

func IoRead(address uint16) byte {
	switch address {
	case 0xFF01:
		return serialData[0]
	case 0xFF02:
		return serialData[1]
	case 0xFF0F:
		return CpuGetIntFlags()
	case 0xFF44:
		return ly
	default:
		if Between16(address, 0xFF04, 0xFF07) {
			return TimerCtx().Read(address)
		}
		logger.Info("UNSUPPORTED bus_read(%04X)\n", address)
		return 0xFF
	}
}

func IoWrite(address uint16, value byte) {
	switch address {
	case 0xFF01:
		serialData[0] = value
	case 0xFF02:
		serialData[1] = value
	case 0xFF0F:
		CpuSetIntFlags(value)
	case 0xFF46:
		RestartDMAContext(value)
		logger.Info("DMA START!")
	default:
		if Between16(address, 0xFF04, 0xFF07) {
			TimerCtx().Write(address, value)
		} else {
			logger.Info("UNSUPPORTED bus_write(%04X)\n", address)
		}
	}
}
