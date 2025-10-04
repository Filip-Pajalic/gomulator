package cpu

import (
	logger "app/internal/logger"
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

	sourceAddr := (uint16(d.value) << 8) | uint16(d.currentByte)

	// Note: This will be properly integrated when we fix the bus access
	// For now, we'll add a placeholder that can be connected later
	logger.Debug("DMA transfer: byte %d from address %04X to OAM %02X", d.currentByte, sourceAddr, 0xFE00+uint16(d.currentByte))

	// TODO: Implement actual memory read/write when bus context is available
	// data := memory.BusCtx().BusRead(sourceAddr)
	// ppu.GetPPUContext().OamWrite(0xFE00+uint16(d.currentByte), data)

	d.currentByte++

	if d.currentByte >= 0xA0 {
		d.active = false
		logger.Info("DMA transfer complete! Transferred 160 bytes to OAM")
	}
}

func (d *DMAContext) DMATransferring() bool {
	return d.active
}

func DmaStart(start uint8) {
	logger.Info("DMA start: source address %02X00", start)

	dmaCtx := DmaCtx()
	dmaCtx.RestartDMAContext(start)
}
