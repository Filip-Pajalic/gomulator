package ui

import (
	"app/internal/cpu"
	"app/internal/input"
	"app/internal/logger"
	"app/internal/memory"
	"time"
)

/*
  Emulator Components:
  | Cartridge (Memory) |
  | CPU                |
  | DMA                |
  | PPU                |
  | Timer              |
  | UI (Optional)      |
*/

// EmuContext holds the state and components of the emulator
type EmuContext struct {
	Paused   bool
	Running  bool
	Ticks    uint64
	Die      bool
	CpuCtx   cpu.CPU
	CartCtx  memory.Cartridge
	PpuCtx   PPU
	timerCtx *cpu.TimerContext
	dmaCtx   cpu.DMA
}

var emuInstance *EmuContext

func EmuCtx(cpuCtx cpu.CPU, cartCtx memory.Cartridge, timerCtx *cpu.TimerContext, dmaCtx cpu.DMA, ppuCtx PPU) *EmuContext {
	return &EmuContext{
		Paused:   false,
		Running:  true,
		Ticks:    0,
		Die:      false,
		CpuCtx:   cpuCtx,
		CartCtx:  cartCtx,
		dmaCtx:   dmaCtx,
		timerCtx: timerCtx,
		PpuCtx:   ppuCtx,
	}
}

// Start begins the emulation by setting flags and initiating the CPU loop
func (e *EmuContext) Start() {
	e.runCPULoop()
}

// runCPULoop continuously steps the CPU until the emulator is stopped
func (e *EmuContext) runCPULoop() {
	for e.Running {
		if e.Paused {
			e.DelayExecution(10)
			continue
		}

		if !e.CpuCtx.Step() {
			e.Die = true
			logger.Fatal("CPU has stopped unexpectedly.")
		}
	}
}

// ExecuteCycles processes a given number of CPU cycles and handles DMA ticks
func (e *EmuContext) ExecuteCycles(cpuCycles int) {
	for i := 0; i < cpuCycles; i++ {
		for n := 0; n < 4; n++ { // Assuming 4 ticks per cycle
			e.Ticks++
			e.timerCtx.Tick()
		}
		e.dmaCtx.DMATick() // Handle DMA operations
	}
}

// DelayExecution pauses execution for the specified milliseconds
func (e *EmuContext) DelayExecution(ms uint32) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// DelaySDL pauses execution using SDL's delay (if needed elsewhere)
func (e *EmuContext) DelaySDL(ms uint32) {
	//sdl.Delay(ms)
}

// LoadROM loads a ROM file into the cartridge context
func (e *EmuContext) LoadROM(romFile string) bool {
	if !e.CartCtx.CartLoad(romFile) {
		logger.Error("Failed to load ROM file:", romFile)
		return false
	}
	return true
}

// StartEmulator initializes all components, loads the ROM, and starts the emulation
func StartEmulator(romFile string) *EmuContext {
	cartContext := memory.CartCtx()

	if !cartContext.CartLoad(romFile) {
		logger.Fatal("ROM loading failed. Exiting emulator.")
	}

	// Continue initializing other components
	timerContext := cpu.TimerCtx()
	dmaContext := cpu.DmaCtx()
	ppuContext := PpuCtx()
	ramContext := memory.RamCtx()
	ioContext := input.NewIo(nil, timerContext, dmaContext)

	busContext := memory.NewBus(cartContext, ramContext, dmaContext, ppuContext, ioContext)
	cpuContext := cpu.NewCpuContext(busContext)

	// Create the emulator context
	emuInstance = EmuCtx(cpuContext, cartContext, timerContext, dmaContext, ppuContext)
	return emuInstance
	//emuInstance.Start()

}
