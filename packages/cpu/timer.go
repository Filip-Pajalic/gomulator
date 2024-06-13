package cpu

type timerContext struct {
	div  uint16
	tima byte
	tma  byte
	tac  byte
}

var ctx = timerContext{0, 0, 0, 0}

func GetTimerContext() *timerContext {
	return &ctx
}

func TimerInit() {
	ctx.div = 0xAC00
}

func TimerTick() {
	prev_div := ctx.div
	ctx.div++

	var timer_update bool

	switch ctx.tac & 0b11 {
	case 0b00:
		timer_update = (prev_div&(1<<9) != 0) && (ctx.div&(1<<9) == 0)
	case 0b01:
		timer_update = (prev_div&(1<<3) != 0) && (ctx.div&(1<<3) == 0)
	case 0b10:
		timer_update = (prev_div&(1<<5) != 0) && (ctx.div&(1<<5) == 0)
	case 0b11:
		timer_update = (prev_div&(1<<7) != 0) && (ctx.div&(1<<7) == 0)
	}

	if timer_update && (ctx.tac&(1<<2) != 0) {
		ctx.tima++
		if ctx.tima == 0 {
			ctx.tima = ctx.tma
			CpuRequestInterrupt(IT_TIMER)
		}
	}
}

func TimerWrite(address uint16, value byte) {
	switch address {
	case 0xFF04:
		// DIV
		ctx.div = 0
	case 0xFF05:
		// TIMA
		ctx.tima = value
	case 0xFF06:
		// TMA
		ctx.tma = value
	case 0xFF07:
		// TAC
		ctx.tac = value
	}
}

func TimerRead(address uint16) byte {
	switch address {
	case 0xFF04:
		return byte(ctx.div >> 8)
	case 0xFF05:
		return ctx.tima
	case 0xFF06:
		return ctx.tma
	case 0xFF07:
		return ctx.tac
	}
	return 0
}
