package cpu

import (
	logger "pajalic.go.emulator/packages/logger"
)

type DMA interface {
	DMATick()
	DMATransferring() bool
}

var dmaInstance *DMAContext

type DMAContext struct {
	active      bool
	currentByte byte
	value       byte
	startDelay  byte
}

func NewDMAContext(start byte) *DMAContext {
	return &DMAContext{
		active:      true,
		currentByte: 0,
		value:       start,
		startDelay:  2,
	}
}

func DmaCtx() *DMAContext {
	if dmaInstance == nil {
		dmaInstance = NewDMAContext(0)
	}
	return dmaInstance
}

func (d *DMAContext) RestartDMAContext(start byte) {
	dmaInstance = NewDMAContext(start)
}

func (d *DMAContext) DMATick() {
	if !d.active {
		return
	}
	if d.startDelay > 0 {
		d.startDelay--
		return
	}
	//Restore this later
	// Calculate source address
	//sourceAddr := (uint16(d.value) << 8) | uint16(d.currentByte)
	// Read data from source
	//data := memory.BusCtx().BusRead(sourceAddr)
	// Write data to OAM
	//ppu.GetPPUContext().OamWrite(uint16(d.currentByte), data)

	d.currentByte++

	// Check if DMA transfer is complete
	if d.currentByte >= 0xA0 {
		d.active = false
		logger.Info("DMA DONE!")
	}
}

func (d *DMAContext) DMATransferring() bool {
	return d.active
}
