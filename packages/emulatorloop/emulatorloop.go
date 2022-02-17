package emulatorloop

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"pajalic.go.emulator/packages/cartridge"
	"pajalic.go.emulator/packages/cpu"
	emu "pajalic.go.emulator/packages/emulator"
	log "pajalic.go.emulator/packages/logger"
)

func Run(argc int, argv []string) int {
	if len(argv) < 2 {
		log.Error("Usage: emu <rom_file>")
	}

	if !cartridge.CartLoad(argv[1]) {
		log.Info("Failed to load ROM file:")
		log.Error("cartridge: ", argv[1])

	}

	log.Info("Cart loaded..")
	sdl.Init(sdl.INIT_VIDEO)
	log.Info("SDL INIT")
	ttf.Init()
	log.Info("TTF INIT")

	emu.GetEmuContext().Running = true
	emu.GetEmuContext().Paused = false
	emu.GetEmuContext().Ticks = 0

	cpu.CpuInit()
	cpu.InitInstructions()

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

	return 0
}

func Delay(ms uint32) {
	sdl.Delay(ms)
}
