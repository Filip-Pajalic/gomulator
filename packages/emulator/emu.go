package emulator

import (
	"pajalic.go.emulator/packages/cpu"
	"pajalic.go.emulator/packages/logger"
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
}

var instance *EmuContext

func NewEmuContext(cpuCtx cpu.CPU, ppuCtx ppu.PPU, dmaCtx cpu.DMA) *EmuContext {
	return &EmuContext{
		Paused:  false,
		Running: false,
		Ticks:   0,
		Die:     false,
		cpuCtx:  cpuCtx,
		ppuCtx:  ppuCtx,
		dmaCtx:  dmaCtx,
	}
}

func GetEmulatorContext() *EmuContext {
	if instance == nil {
		cpuContext := cpu.GetCpuContext()
		ppuContext := ppu.GetPPUContext()
		dmaContext := cpu.GetDMAContext()
		instance = NewEmuContext(cpuContext, ppuContext, dmaContext)
	}
	return instance
}

func (e *EmuContext) Initialize() {
	//e.cpu.InitializeTimer()
	//e.cpu.LoadInstructions()
	//e.ppu.Initialize()
	//input.InitializeGamePad()
}

func (e *EmuContext) Start() {
	e.Initialize()

	e.Running = true
	e.Paused = false
	e.Ticks = 0

	go e.runCPULoop()
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
