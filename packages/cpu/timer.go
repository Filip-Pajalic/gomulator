package cpu

type Timer interface {
	Tick(address uint16, value byte)
	Write(address uint16, value byte)
	Read(address uint16) byte
}

type TimerContext struct {
	div  uint16
	tima byte
	tma  byte
	tac  byte
}

var timerInstance *TimerContext

func GetTimerContext() *TimerContext {
	if timerInstance != nil {
		timerInstance = &TimerContext{
			div: 0xAC00,
		}
	}
	return timerInstance
}

func (t *TimerContext) Tick() {
	prev_div := t.div
	t.div++

	var timer_update bool

	switch t.tac & 0b11 {
	case 0b00:
		timer_update = (prev_div&(1<<9) != 0) && (t.div&(1<<9) == 0)
	case 0b01:
		timer_update = (prev_div&(1<<3) != 0) && (t.div&(1<<3) == 0)
	case 0b10:
		timer_update = (prev_div&(1<<5) != 0) && (t.div&(1<<5) == 0)
	case 0b11:
		timer_update = (prev_div&(1<<7) != 0) && (t.div&(1<<7) == 0)
	}

	if timer_update && (t.tac&(1<<2) != 0) {
		t.tima++
		if t.tima == 0 {
			t.tima = t.tma
			//CpuRequestInterrupt(IT_TIMER)
		}
	}
}

func (t *TimerContext) TimerWrite(address uint16, value byte) {
	switch address {
	case 0xFF04:
		// DIV
		t.div = 0
	case 0xFF05:
		// TIMA
		t.tima = value
	case 0xFF06:
		// TMA
		t.tma = value
	case 0xFF07:
		// TAC
		t.tac = value
	}
}

func (t *TimerContext) TimerRead(address uint16) byte {
	switch address {
	case 0xFF04:
		return byte(t.div >> 8)
	case 0xFF05:
		return t.tima
	case 0xFF06:
		return t.tma
	case 0xFF07:
		return t.tac
	}
	return 0
}
