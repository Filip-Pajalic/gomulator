package cpu

type CycleManager struct {
	ticks          int32
	peripheralTick func(tCycles int32)
}

var Cm = &CycleManager{}

func (c *CycleManager) IncreaseCycle(tickAmount int32) {
	c.ticks += tickAmount

	// Batch advance the timer for better performance
	if tickAmount > 0 {
		timer := TimerCtx()
		totalTicks := tickAmount * 4
		timer.TickBatch(totalTicks)
		if c.peripheralTick != nil {
			c.peripheralTick(totalTicks)
		}
	}
}

func (c *CycleManager) GetCycleTicks() int32 {
	return c.ticks
}

func (c *CycleManager) SetPeripheralTick(fn func(tCycles int32)) {
	c.peripheralTick = fn
}
