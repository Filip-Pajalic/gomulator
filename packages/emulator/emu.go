// emu.go
package emulator

import (
	"context"
	"github.com/veandco/go-sdl2/sdl"
	"pajalic.go.emulator/packages/cpu"
	"pajalic.go.emulator/packages/logger"
	"pajalic.go.emulator/packages/memory"
	"pajalic.go.emulator/packages/ppu"
	"sync"
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
	Paused  bool
	Running bool
	Ticks   uint64
	Die     bool
	cpuCtx  cpu.CPU
	ppuCtx  *ppu.PpuContext // Using concrete type for consistency
	dmaCtx  cpu.DMA
	cartCtx memory.Cartridge

	mu sync.RWMutex // Mutex to protect emulator flags

	wg     sync.WaitGroup     // WaitGroup to manage goroutines
	ctx    context.Context    // Context for cancellation
	cancel context.CancelFunc // Cancel function for context
}

// Singleton instance of EmuContext
var emuInstance *EmuContext

// newEmuContext initializes a new EmuContext with provided components
func newEmuContext(cpuCtx cpu.CPU, ppuCtx *ppu.PpuContext, dmaCtx cpu.DMA, cartCtx memory.Cartridge) *EmuContext {
	ctx, cancel := context.WithCancel(context.Background())
	return &EmuContext{
		Paused:  false,
		Running: false,
		Ticks:   0,
		Die:     false,
		cpuCtx:  cpuCtx,
		ppuCtx:  ppuCtx,
		dmaCtx:  dmaCtx,
		cartCtx: cartCtx,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start begins the emulation by setting flags and initiating the CPU loop
func (e *EmuContext) Start() {
	e.mu.Lock()
	e.Running = true
	e.Paused = false
	e.Ticks = 0
	e.mu.Unlock()

	// Start the CPU loop in a separate goroutine
	e.wg.Add(1)
	go e.runCPULoop()

	// Start the PPU main loop in a separate goroutine
	e.wg.Add(1)
	go e.ppuCtx.RunPPUMainLoop(e.ctx, &e.wg)
}

// runCPULoop continuously steps the CPU until the emulator is stopped
func (e *EmuContext) runCPULoop() {
	defer e.wg.Done()
	for {
		select {
		case <-e.ctx.Done():
			logger.Info("CPU loop received cancellation signal.")
			return
		default:
			e.mu.RLock()
			if e.Die {
				e.mu.RUnlock()
				e.cancel() // Signal cancellation
				return
			}
			paused := e.Paused
			e.mu.RUnlock()

			if paused {
				e.DelayExecution(10)
				continue
			}

			// Step the CPU; if it returns false, stop the emulator
			if !e.cpuCtx.Step() {
				e.mu.Lock()
				e.Die = true
				e.mu.Unlock()
				logger.Fatal("CPU has stopped unexpectedly.")
				e.cancel() // Signal cancellation
				return
			}

			// Execute additional cycles if needed
			// e.ExecuteCycles(1) // Uncomment if cycle-based execution is required
		}
	}
}

// ExecuteCycles processes a given number of CPU cycles and handles DMA ticks
func (e *EmuContext) ExecuteCycles(cpuCycles int) {
	for i := 0; i < cpuCycles; i++ {
		for n := 0; n < 4; n++ { // Assuming 4 ticks per cycle
			e.mu.Lock()
			e.Ticks++
			e.mu.Unlock()
			// e.cpuCtx.TimerTick() // Uncomment if TimerTick is implemented
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
	// Initialize CPU, PPU, DMA, and Cartridge contexts
	cpuContext := cpu.CpuCtx()
	ppuContext := ppu.PpuCtx()
	dmaContext := cpu.GetDMAContext()
	cartContext := memory.CartCtx()

	// Create the emulator context
	emuInstance = newEmuContext(cpuContext, ppuContext, dmaContext, cartContext)

	// Initialize Memory
	memory.InitializeMemory()

	// Load the ROM; if failed, terminate the emulator
	if !emuInstance.LoadROM(romFile) {
		logger.Fatal("ROM loading failed. Exiting emulator.")
	}

	// Start the PPU component to handle VRAM and OAM events
	ppuContext.StartPPUComponent()

	// Start the emulator (CPU loop and other components)
	emuInstance.Start()

	// Main loop to keep the emulator running until it needs to stop
	emuInstance.wg.Add(1)
	go func() {
		defer emuInstance.wg.Done()
		for {
			select {
			case <-emuInstance.ctx.Done():
				logger.Info("Main emulator loop received cancellation signal.")
				return
			default:
				emuInstance.mu.RLock()
				if emuInstance.Die {
					emuInstance.mu.RUnlock()
					emuInstance.cancel() // Signal cancellation
					return
				}
				emuInstance.mu.RUnlock()

				// Implement any additional main loop logic here
				// For example, updating the UI, handling user input, etc.

				// To prevent the loop from consuming 100% CPU, sleep briefly
				time.Sleep(time.Millisecond * 1)
			}
		}
	}()

	// Wait until all goroutines have finished
	emuInstance.wg.Wait()

	// Cleanup actions after the emulator stops
	// For example, saving state, etc.
}
