package cpu

import (
	"app/internal/logger"
)

type Timer interface {
	Tick()
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

func TimerCtx() *TimerContext {
	if timerInstance == nil {
		timerInstance = &TimerContext{
			div: 0xAC00,
		}
	}
	return timerInstance
}

func (t *TimerContext) Tick() {
	prevDiv := t.div
	t.div++

	if t.tac&(1<<2) != 0 { // Timer enabled
		var timerUpdate bool
		switch t.tac & 0x03 {
		case 0x00:
			// 4096 Hz (Bit 9)
			timerUpdate = (prevDiv&(1<<9) != 0) && (t.div&(1<<9) == 0)
		case 0x01:
			// 262144 Hz (Bit 3)
			timerUpdate = (prevDiv&(1<<3) != 0) && (t.div&(1<<3) == 0)
		case 0x02:
			// 65536 Hz (Bit 5)
			timerUpdate = (prevDiv&(1<<5) != 0) && (t.div&(1<<5) == 0)
		case 0x03:
			// 16384 Hz (Bit 7)
			timerUpdate = (prevDiv&(1<<7) != 0) && (t.div&(1<<7) == 0)
		}

		if timerUpdate {
			prevTima := t.tima
			t.tima++
			if t.tima == 0 {
				t.tima = t.tma
				logger.Debug("Timer overflow: prev=%02X reload=%02X div=%04X tac=%02X", prevTima, t.tma, t.div, t.tac)
				CpuCtx().RequestInterrupt(IT_TIMER)
			}
		}
	}
}

// TickBatch processes multiple timer ticks at once for better performance
func (t *TimerContext) TickBatch(ticks int32) {
	if ticks <= 0 {
		return
	}

	// Fast path: if timer is disabled, just update div
	if t.tac&(1<<2) == 0 {
		t.div += uint16(ticks)
		return
	}

	// Timer is enabled - need to check for overflows
	// We still need to process each tick to detect falling edges correctly
	for i := int32(0); i < ticks; i++ {
		t.Tick()
	}
}

func (t *TimerContext) Write(address uint16, value byte) {
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
		t.tac = value & 0x07 // Only the lower 3 bits are used
	}
}

func (t *TimerContext) Read(address uint16) byte {
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
	return 0xFF
}
