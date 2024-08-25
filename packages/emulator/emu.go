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
  Emu components:
  |Cart|
  |CPU|
  |Address Bus|
  |PPU|
  |Timer|
*/

type EmuContext struct {
	Paused  bool
	Running bool
	Ticks   uint64
	Die     bool
	cpuCtx  cpu.CPU
	ppuCtx  ppu.PPU
	dmaCtx  cpu.DMA
	cartCtx memory.Cartridge
}

var emuInstance *EmuContext

func newEmuContext(cpuCtx cpu.CPU, ppuCtx ppu.PPU, dmaCtx cpu.DMA, cartCtx memory.Cartridge) *EmuContext {

	return &EmuContext{
		Paused:  false,
		Running: false,
		Ticks:   0,
		Die:     false,
		cpuCtx:  cpuCtx,
		ppuCtx:  ppuCtx,
		dmaCtx:  dmaCtx,
		cartCtx: cartCtx,
	}
}

func (e *EmuContext) Start() {
	e.Running = true

	e.Paused = false
	e.Ticks = 0

	e.runCPULoop()
	//e.runUI()
}

func (e *EmuContext) runCPULoop() {
	for e.Running {

		if e.Paused {
			DelayExecution(10)
			continue
		}

		if !e.cpuCtx.Step() {
			e.Die = true
			logger.Fatal("CPU Stopped")
		}
	}
}

func (e *EmuContext) ExecuteCycles(cpuCycles int) {
	for i := 0; i < cpuCycles; i++ {
		for n := 0; n < 4; n++ {
			e.Ticks++
			//e.cpuCtx.TimerTick()
		}
		e.dmaCtx.DMATick()
	}
}

func (e *EmuContext) runui() {
	/*	ui.initialize()

		for !e.die {
			time.sleep(1000)
			ui.handleevents()
			ui.update()
		}
		ui.destroy()*/
}

func DelayExecution(ms uint32) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func Delay(ms uint32) {
	sdl.Delay(ms)
}

func (e *EmuContext) LoadROM(romFile string) bool {

	if !e.cartCtx.CartLoad(romFile) {
		logger.Error("Failed to load ROM file:", romFile)
		return false
	}
	return true
}

func StartEmulator(romFile string) {
	cpuContext := cpu.CpuCtx()
	ppuContext := ppu.PpuCtx()
	dmaContext := cpu.GetDMAContext()
	cartContext := memory.CartCtx()
	emuInstance = newEmuContext(cpuContext, ppuContext, dmaContext, cartContext)
	emuInstance.LoadROM(romFile)
	memory.SubscribeLoop()
	emuInstance.Start()

}
