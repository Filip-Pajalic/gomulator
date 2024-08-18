package emulator

import (
	"github.com/veandco/go-sdl2/sdl"
	"pajalic.go.emulator/packages/logger"
	"pajalic.go.emulator/packages/memory"
)

/*func CpuRun() {
	cpu.getcpucontext().cpuinit()
	cpu.TimerInit()
	cpu.InitInstructions()
	ppu.PpuInit()
	input.GamePadInit()
	ppu.PpuInit()

	GetEmuContext().Running = true
	GetEmuContext().Paused = false
	GetEmuContext().Ticks = 0

	for GetEmuContext().Running {
		if GetEmuContext().Paused {
			Delay(10)
			continue
		}

		if !cpu.CpuStep() {
			GetEmuContext().Die = true
			log.Fatal("CPU Stopped")
		}
	}
}

func Run(argc int, argv []string) int {
	if len(argv) < 2 {
		log.Error("Usage: emu <rom_file>")
	}

	if !memory.CartLoad(argv[1]) {
		log.Info("Failed to load ROM file:")
		log.Error("cartridge: ", argv[1])

	}

	GetEmuContext().Die = false
	go CpuRun()

	//previousFrame := cpu.PpuCtx.CurrentFrame
	ui.UiInit()

	for !GetEmuContext().Die {

		//same as usleep?
		time.Sleep(1000)
		ui.UiHandleEvents()
		//if previousFrame != cpu.PpuCtx.CurrentFrame {
		ui.UiUpdate()
		//}
		//	previousFrame = cpu.PpuCtx.CurrentFrame
	}
	ui.DestroyWindow()

	return 0

}*/

func Delay(ms uint32) {
	sdl.Delay(ms)
}

func LoadROM(romFile string) bool {
	if !memory.CartLoad(romFile) {
		logger.Error("Failed to load ROM file:", romFile)
		return false
	}
	return true
}

func StartEmulator() {
	GetEmulatorContext().Start()
}
