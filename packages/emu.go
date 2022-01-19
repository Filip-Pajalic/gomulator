package gameboypackage

import (
	"fmt"
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

var Etx emuContext

type emuContext struct {
	Paused  bool
	Running bool
	Ticks   uint64
}

func Emu_run(argc int, argv []string) int {
	if len(argv) < 2 {
		fmt.Printf("Usage: emu <rom_file>\n")
		return -1
	}

	if !cartLoad(argv[1]) {
		fmt.Printf("Failed to load ROM file: %s\n", argv[1])
		return -2
	}
	fmt.Printf("Cart loaded..\n")

	sdl.Init(sdl.INIT_VIDEO)
	fmt.Printf("SDL INIT\n")
	ttf.Init()
	fmt.Printf("TTF INIT\n")

	CpuInit()

	Etx.Running = true
	Etx.Paused = false
	Etx.Ticks = 0

	for Etx.Running {
		if Etx.Paused {
			delay(10)
			continue
		}

		if !CpuStep() {
			fmt.Printf("CPU Stopped\n")
			return -3
		}

		Etx.Ticks++
	}

	return 0
}

func getEmuContext() *emuContext {
	return &Etx
}

func delay(ms uint32) {
	sdl.Delay(ms)
}
