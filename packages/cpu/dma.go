package cpu

import (
	"pajalic.go.emulator/packages/memory"
	"pajalic.go.emulator/packages/ppu"
	"time"

	log "pajalic.go.emulator/packages/logger"
)

type DMA interface {
	DMATick()
	DMATransferring() bool
}

var instance *DMAContext

type DMAContext struct {
	active     bool
	byte       byte
	value      byte
	startDelay byte
}

func NewDmaContext(start byte) *DMAContext {
	return &DMAContext{
		active:     true,
		byte:       0,
		value:      start,
		startDelay: 2,
	}
}

func GetDMAContext() *DMAContext {
	if instance == nil {
		instance = NewDmaContext(2)
	}
	return instance
}

func RestartDMAContext(start byte) *DMAContext {
	instance = NewDmaContext(start)
	return instance
}

/*
	func DmaStart(start byte) {
		dmaCtx.active = true
		dmaCtx.byte = 0
		dmaCtx.startDelay = 2
		dmaCtx.value = start
	}
*/
func (d *DMAContext) DMATick() {
	if !d.active {
		return
	}
	//is it bigger than zero or equals to 1 here
	if d.startDelay > 0 {
		d.startDelay--
		return
	}
	//might be wrong
	ppu.PpuOamWrite(uint16(d.byte), memory.BusRead((uint16(d.value)*0x100)+(uint16(d.byte))))

	d.byte++

	d.active = d.byte < 0xA0

	if !d.active {
		log.Info("DMA DONE!\n")
		time.Sleep(2)
	}
}

func (d *DMAContext) DMATransferring() bool {
	return d.active
}
