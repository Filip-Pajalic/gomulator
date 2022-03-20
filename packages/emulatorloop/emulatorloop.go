package emulatorloop

import (
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"pajalic.go.emulator/packages/cartridge"
	"pajalic.go.emulator/packages/cpu"
	emu "pajalic.go.emulator/packages/emulator"
	log "pajalic.go.emulator/packages/logger"
	"pajalic.go.emulator/packages/ui"
)

func CpuRun() {
	cpu.CpuInit()

	emu.GetEmuContext().Running = true
	emu.GetEmuContext().Paused = false
	emu.GetEmuContext().Ticks = 0

	for emu.GetEmuContext().Running {
		if emu.GetEmuContext().Paused {
			Delay(10)
			continue
		}

		if !cpu.CpuStep() {
			log.Error("CPU Stopped")
		}

		emu.GetEmuContext().Ticks++
	}
}

func Run(argc int, argv []string) int {
	if len(argv) < 2 {
		log.Error("Usage: emu <rom_file>")
	}

	if !cartridge.CartLoad(argv[1]) {
		log.Info("Failed to load ROM file:")
		log.Error("cartridge: ", argv[1])

	}

	ui.UiInit()

	go CpuRun()

	for !emu.GetEmuContext().Die {
		//same as usleep?
		time.Sleep(1000)
		ui.UiHandleEvents()
	}

	cpu.InitInstructions()

	return 0

}

func Delay(ms uint32) {
	sdl.Delay(ms)
}
