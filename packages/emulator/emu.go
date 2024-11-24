package emulator

import (
	"github.com/veandco/go-sdl2/sdl"
	"pajalic.go.emulator/packages/cpu"
	"pajalic.go.emulator/packages/logger"
	"pajalic.go.emulator/packages/memory"
	"pajalic.go.emulator/packages/ppu"
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
	cpuCtx   cpu.CPU
	cartCtx  memory.Cartridge
	timerCtx *cpu.TimerContext
	dmaCtx   cpu.DMA
}

var emuInstance *EmuContext

func EmuCtx(cpuCtx cpu.CPU, cartCtx memory.Cartridge, timerCtx *cpu.TimerContext, dmaCtx cpu.DMA) *EmuContext {
	return &EmuContext{
		Paused:   false,
		Running:  true,
		Ticks:    0,
		Die:      false,
		cpuCtx:   cpuCtx,
		cartCtx:  cartCtx,
		dmaCtx:   dmaCtx,
		timerCtx: timerCtx,
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

		if !e.cpuCtx.Step() {
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
	sdl.Delay(ms)
}

// LoadROM loads a ROM file into the cartridge context
func (e *EmuContext) LoadROM(romFile string) bool {
	if !e.cartCtx.CartLoad(romFile) {
		logger.Error("Failed to load ROM file:", romFile)
		return false
	}
	return true
}

// StartEmulator initializes all components, loads the ROM, and starts the emulation
func StartEmulator(romFile string) {
	cartContext := memory.CartCtx()

	if !cartContext.CartLoad(romFile) {
		logger.Fatal("ROM loading failed. Exiting emulator.")
	}

	// Continue initializing other components
	timerContext := cpu.TimerCtx()
	dmaContext := cpu.DmaCtx()
	ppuContext := ppu.PpuCtx()
	ramContext := memory.RamCtx()

	busContext := memory.NewBus(cartContext, ramContext, dmaContext, ppuContext)
	cpuContext := cpu.NewCpuContext(busContext)

	// Create the emulator context
	emuInstance = EmuCtx(cpuContext, cartContext, timerContext, dmaContext)

	emuInstance.Start()
}
