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

	var prev_div = ctx.div

	var prev_div_bool = false

	ctx.div++

	var div_bool = false

	var timer_update = false

	switch ctx.tac & (0b11) {
	case 0b00:
		if (prev_div & (1 << 9)) == 1 {
			prev_div_bool = true
		}

		if !(ctx.div&(1<<9) == 1) {
			div_bool = true
		}

		timer_update = prev_div_bool && div_bool
		break
	case 0b01:
		if (prev_div & (1 << 3)) == 1 {
			prev_div_bool = true
		}

		if !(ctx.div&(1<<3) == 1) {
			div_bool = true
		}

		timer_update = prev_div_bool && div_bool
		break
	case 0b10:
		if (prev_div & (1 << 5)) == 1 {
			prev_div_bool = true
		}

		if !(ctx.div&(1<<5) == 1) {
			div_bool = true
		}

		timer_update = prev_div_bool && div_bool
		break
	case 0b11:
		if (prev_div & (1 << 7)) == 1 {
			prev_div_bool = true
		}

		if !(ctx.div&(1<<7) == 1) {
			div_bool = true
		}

		timer_update = prev_div_bool && div_bool
		break
	}

	var tac_bool = false
	if ctx.tac&(1<<2) == 1 {
		prev_div_bool = true
	}
	if timer_update && tac_bool {
		ctx.tima++

		if ctx.tima == 0xFF {
			ctx.tima = ctx.tma

			CpuRequestInterrupt(IT_TIMER)
		}
	}

}

func TimerWrite(address uint16, value byte) {
	switch address {
	case 0xFF04:
		//DIV
		ctx.div = 0
		break

	case 0xFF05:
		//TIMA
		ctx.tima = value
		break

	case 0xFF06:
		//TMA
		ctx.tma = value
		break

	case 0xFF07:
		//TAC
		ctx.tac = value
		break
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
