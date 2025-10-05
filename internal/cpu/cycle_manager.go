package cpu

type CycleManager struct {
	ticks int32
}

var Cm = &CycleManager{}

func (c *CycleManager) IncreaseCycle(tickAmount int32) {
	c.ticks += tickAmount

	// Advance the timer for each CPU clock cycle (4 clocks per machine cycle)
	if tickAmount > 0 {
		timer := TimerCtx()
		totalTicks := tickAmount * 4
		for i := int32(0); i < totalTicks; i++ {
			timer.Tick()
		}
	}
}

func (c *CycleManager) GetCycleTicks() int32 {
	return c.ticks
}
