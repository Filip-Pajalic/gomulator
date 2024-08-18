package cpu

import (
	log "pajalic.go.emulator/packages/logger"
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
			return timerInstance.TimerRead(address)
		}
		log.Warn("UNSUPPORTED bus_read(%04X)\n", address)
		return 0
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
		log.Info("DMA START!\n")
	default:
		if Between16(address, 0xFF04, 0xFF07) {
			timerInstance.TimerWrite(address, value)
		} else {
			log.Warn("UNSUPPORTED bus_write(%04X)\n", address)
		}
	}
}
