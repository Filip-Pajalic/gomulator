package cpu

/*
  Emu components:
  |Cart|
  |CPU|
  |Address Bus|
  |PPU|
  |Timer|
*/

/*var log = logrus.New()

func init() {
	log.Out = os.Stdout
	log.Formatter = &logrus.TextFormatter{
		DisableTimestamp: true,
	}
	log.Level = logrus.DebugLevel
}*/

var ectx emuContext

type emuContext struct {
	Paused  bool
	Running bool
	Ticks   uint64
	Die     bool
}

func EmuCycles(cpuCycles int) {
	for i := 0; i < cpuCycles; i++ {
		for n := 0; n < 4; n++ {
			ectx.Ticks++
			TimerTick()
		}

		dmaTick()
	}
}

func GetEmuContext() *emuContext {
	return &ectx
}
