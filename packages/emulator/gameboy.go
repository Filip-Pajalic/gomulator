package emulator

/*func CpuRun() {




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
