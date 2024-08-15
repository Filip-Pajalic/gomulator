package emulator

import "pajalic.go.emulator/packages/cpu"

/*
  Emu components:
  |Cart|
  |CPU|
  |Address Bus|
  |PPU|
  |Timer|
*/

var instance *EmuContext

type EmuContext struct {
	Paused  bool
	Running bool
	Ticks   uint64
	Die     bool
	dma     cpu.DMA
}

func NewEmuContext(dma cpu.DMA) *EmuContext {
	return &EmuContext{
		Paused:  false,
		Running: false,
		Ticks:   0,
		Die:     false,
		dma:     dma,
	}
}

func GetEmuContext() *EmuContext {
	if instance == nil {
		dmaContext := cpu.GetDMAContext()
		instance = NewEmuContext(dmaContext)
	}
	return instance
}

func (e *EmuContext) EmuCycles(cpuCycles int) {
	for i := 0; i < cpuCycles; i++ {
		for n := 0; n < 4; n++ {
			e.Ticks++
			// Assume cpu.TimerTick() and cpu.dmaTick() are defined elsewhere
			cpu.TimerTick()
		}
		e.dma.DMATick()
	}
}
