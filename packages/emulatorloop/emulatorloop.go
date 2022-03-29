package emulatorloop

import (
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"pajalic.go.emulator/packages/cpu"
	log "pajalic.go.emulator/packages/logger"
	"pajalic.go.emulator/packages/ui"
)

func CpuRun() {
	cpu.CpuInit()
	cpu.TimerInit()
	cpu.InitInstructions()
	cpu.PpuInit()

	cpu.GetEmuContext().Running = true
	cpu.GetEmuContext().Paused = false
	cpu.GetEmuContext().Ticks = 0

	for cpu.GetEmuContext().Running {
		if cpu.GetEmuContext().Paused {
			Delay(10)
			continue
		}

		if !cpu.CpuStep() {
			cpu.GetEmuContext().Die = true
			log.Fatal("CPU Stopped")
		}
	}
}

func Run(argc int, argv []string) int {
	if len(argv) < 2 {
		log.Error("Usage: emu <rom_file>")
	}

	if !cpu.CartLoad(argv[1]) {
		log.Info("Failed to load ROM file:")
		log.Error("cartridge: ", argv[1])

	}

	ui.UiInit()
	cpu.GetEmuContext().Die = false
	go CpuRun()

	for !cpu.GetEmuContext().Die {
		//same as usleep?
		time.Sleep(1000)
		ui.UiHandleEvents()
		ui.UiUpdate()
	}
	ui.DestroyWindow()

	return 0

}

func Delay(ms uint32) {
	sdl.Delay(ms)
}
