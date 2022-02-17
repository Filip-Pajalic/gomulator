package emulator

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

var ctx emuContext

type emuContext struct {
	Paused  bool
	Running bool
	Ticks   uint64
}

func EmuCycles(cpuCycles int) {

}

func GetEmuContext() *emuContext {
	return &ctx
}
