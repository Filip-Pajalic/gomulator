package cpu

type CycleManager struct {
	ticks int32
}

var Cm *CycleManager

func (c *CycleManager) IncreaseCycle(tickAmount int32) {
	c.ticks += tickAmount
}

func (c *CycleManager) GetCycleTicks() int32 {
	return c.ticks
}
