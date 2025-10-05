package cpu

import (
	logger "app/internal/logger"
	"app/internal/memory"
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
	if d == nil {
		dmaInstance = NewDMAContext(start)
		return
	}

	d.active = true
	d.currentByte = 0
	d.value = start
	d.startDelay = 2
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
	destAddr := 0xFE00 + uint16(d.currentByte)

	data := memory.BusCtx().BusRead(sourceAddr)
	memory.BusCtx().DmaWriteToOam(destAddr, data)

	logger.Debug("DMA transfer: byte %d from %04X -> %04X data=%02X", d.currentByte, sourceAddr, destAddr, data)

	d.currentByte++

	if d.currentByte >= 0xA0 {
		d.active = false
		logger.Debug("DMA transfer complete! Transferred 160 bytes to OAM")
	}
}

func (d *DMAContext) DMATransferring() bool {
	return d.active
}

func DmaStart(start uint8) {
	logger.Debug("DMA start: source address %02X00", start)

	dmaCtx := DmaCtx()
	dmaCtx.RestartDMAContext(start)
}
