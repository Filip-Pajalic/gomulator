package gameboypackage

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

/*
  Emu components:
  |Cart|
  |CPU|
  |Address Bus|
  |PPU|
  |Timer|
*/

var Logger = logrus.New()

func init() {
	Logger.Out = os.Stdout
	Logger.Formatter = &logrus.TextFormatter{
		DisableTimestamp: true,
	}
	Logger.Level = logrus.DebugLevel
}

var Etx emuContext

type emuContext struct {
	Paused  bool
	Running bool
	Ticks   uint64
}

func Emu_run(argc int, argv []string) int {
	if len(argv) < 2 {
		Logger.Error("Usage: emu <rom_file>")
	}

	if !cartLoad(argv[1]) {
		Logger.WithFields(logrus.Fields{
			"rom": argv[1],
		}).Error("Failed to load ROM file:")
	}

	Logger.Info("Cart loaded..")
	sdl.Init(sdl.INIT_VIDEO)
	Logger.Info("SDL INIT")
	ttf.Init()
	Logger.Info("TTF INIT")
	CpuInit()
	initInstructions()

	Etx.Running = true
	Etx.Paused = false
	Etx.Ticks = 0

	for Etx.Running {
		if Etx.Paused {
			delay(10)
			continue
		}

		if !CpuStep() {
			Logger.Error("CPU Stopped")
		}

		Etx.Ticks++
	}

	return 0
}

func EmuCycles(cpuCycles int) {

}

func getEmuContext() *emuContext {
	return &Etx
}

func delay(ms uint32) {
	sdl.Delay(ms)
}
