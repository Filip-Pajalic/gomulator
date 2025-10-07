package ui

import (
	"app/internal/cpu"
	"app/internal/input"
	"app/internal/logger"
	"app/internal/memory"
	"errors"
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
	BusCtx   *memory.Bus
}

var emuInstance *EmuContext

var ErrEmulationStopped = errors.New("emulation stopped")

func EmuCtx(cpuCtx cpu.CPU, cartCtx memory.Cartridge, timerCtx *cpu.TimerContext, dmaCtx cpu.DMA, ppuCtx PPU, busCtx *memory.Bus) *EmuContext {
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
		BusCtx:   busCtx,
	}
}

func (e *EmuContext) Start() {
	e.runCPULoop()
}

func (e *EmuContext) runCPULoop() {
	for e.Running {
		if e.Paused {
			e.DelayExecution(10)
			continue
		}

		if !e.CpuCtx.Step() {
			if e.handleCpuStop() {
				break
			}
			return
		}
	}
}

func (e *EmuContext) ExecuteCycles(cpuCycles int) {
	if !e.Running {
		return
	}

	if e.PpuCtx == nil {
		logger.Fatal("PPU Context is nil!")
		return
	}

	// Convert requested CPU cycles to machine cycles (1 machine cycle = 4 CPU cycles)
	remainingMachineCycles := int32((cpuCycles + 3) / 4)
	if remainingMachineCycles <= 0 {
		return
	}

	targetTicks := cpu.Cm.GetCycleTicks() + remainingMachineCycles

	for cpu.Cm.GetCycleTicks() < targetTicks {
		prevTicks := cpu.Cm.GetCycleTicks()

		if !e.CpuCtx.Step() {
			if e.handleCpuStop() {
				return
			}
			return
		}

		consumedTicks := cpu.Cm.GetCycleTicks() - prevTicks
		if consumedTicks <= 0 {
			consumedTicks = 1
		}

		cpuSteps := consumedTicks * 4

		// OPTIMIZED: Batch tick the PPU and DMA instead of looping
		e.PpuCtx.PpuTickBatch(cpuSteps)
		e.dmaCtx.DMATickBatch(cpuSteps)

		e.Ticks += uint64(cpuSteps)
	}
}

// DelayExecution pauses execution for the specified milliseconds
func (e *EmuContext) DelayExecution(ms uint32) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func (e *EmuContext) DelaySDL(ms uint32) {
	//sdl.Delay(ms)
}

func (e *EmuContext) LoadROM(romFile string) bool {
	if !e.CartCtx.CartLoad(romFile) {
		logger.Error("Failed to load ROM file: %s", romFile)
		return false
	}
	return true
}

func (e *EmuContext) StepFrame() {
	// Game Boy frame = 154 lines × 456 cycles per line = 70,224 cycles
	const LINES_PER_FRAME = 154
	const TICKS_PER_LINE = 456
	frameCycles := LINES_PER_FRAME * TICKS_PER_LINE
	logger.Debug("StepFrame: executing %d cycles for complete frame", frameCycles)
	if !e.Running {
		return
	}
	e.ExecuteCycles(frameCycles)
}

func (e *EmuContext) handleCpuStop() bool {
	if e.CpuCtx.IsStopped() {
		if e.Running {
			e.Running = false
			e.Die = false
			logger.Info("CPU STOP encountered; ending emulation loop")
		}
		return true
	}
	e.Die = true
	logger.Debug("CPU has stopped unexpectedly.")
	return false
}

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

	cpuContext := cpu.NewCpuContext(nil) // Bus will be set later
	busContext := memory.NewBus(cartContext, ramContext, dmaContext, ppuContext, ioContext, cpuContext)

	cpuContext = cpu.NewCpuContext(busContext)

	emuInstance = EmuCtx(cpuContext, cartContext, timerContext, dmaContext, ppuContext, busContext)

	return emuInstance

}

// StartEmulatorFromBytes initializes the emulator from a ROM byte slice (for WASM/JS)
func StartEmulatorFromBytes(romBytes []byte) *EmuContext {
	cartContext := memory.CartCtx()
	cartContext.LoadROMFromBytes(romBytes)

	// Continue initializing other components
	timerContext := cpu.TimerCtx()
	dmaContext := cpu.DmaCtx()
	ppuContext := PpuCtx()
	ramContext := memory.RamCtx()

	ioContext := input.NewIo(nil, timerContext, dmaContext)

	cpuContext := cpu.NewCpuContext(nil) // Bus will be set later
	busContext := memory.NewBus(cartContext, ramContext, dmaContext, ppuContext, ioContext, cpuContext)

	cpuContext = cpu.NewCpuContext(busContext)

	emuInstance = EmuCtx(cpuContext, cartContext, timerContext, dmaContext, ppuContext, busContext)

	return emuInstance
}
