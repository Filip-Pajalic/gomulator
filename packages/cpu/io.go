package cpu

import (
	log "pajalic.go.emulator/packages/logger"
)

var serialData [2]byte
var ly byte = 0

func IoRead(address uint16) byte {
	if address == 0xFF01 {
		return serialData[0]
	}

	if address == 0xFF02 {
		return serialData[1]
	}

	if Between16(address, 0xFF04, 0xFF07) {
		return TimerRead(address)
	}

	if address == 0xFF0F {
		return CpuGetIntFlags()
	}

	if address == 0xFF44 {
		//ly++
		return 0
	}

	log.Warn("UNSUPPORTED bus_read(%04X)\n", address)
	return 0
}

func IoWrite(address uint16, value byte) {
	if address == 0xFF01 {
		serialData[0] = value
		return
	}

	if address == 0xFF02 {
		serialData[1] = value
		return
	}
	//potentiall issue
	if Between16(address, 0xFF04, 0xFF07) {
		TimerWrite(address, value)
		return
	}

	if address == 0xFF0F {
		CpuSetIntFlags(value)
		return
	}

	if address == 0xFF46 {
		DmaStart(value)
		log.Info("DMA START!\n")
	}

	log.Warn("UNSUPPORTED bus_write(%04X)\n", address)
}
