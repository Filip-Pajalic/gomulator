package cpu

import (
	"time"

	log "pajalic.go.emulator/packages/logger"
)

type dmaContext struct {
	active     bool
	byte       byte
	value      byte
	startDelay byte
}

var dmaCtx dmaContext

func DmaStart(start byte) {
	dmaCtx.active = true
	dmaCtx.byte = 0
	dmaCtx.startDelay = 2
	dmaCtx.value = start
}

func dmaTick() {
	if !dmaCtx.active {
		return
	}
	//is it bigger than zero or equals to 1 here
	if dmaCtx.startDelay > 0 {
		dmaCtx.startDelay--
		return
	}
	//might be wrong
	PpuOamWrite(uint16(dmaCtx.byte), BusRead((uint16(dmaCtx.value)*0x100)+(uint16(dmaCtx.byte))))

	dmaCtx.byte++

	dmaCtx.active = dmaCtx.byte < 0xA0

	if !dmaCtx.active {
		log.Info("DMA DONE!\n")
		time.Sleep(2)
	}
}

func DmaTransferring() bool {
	return dmaCtx.active
}
