package cpu

type InterruptType byte

const (
	IT_VBLANK   InterruptType = 1
	IT_LCD_STAT               = 2
	IT_TIMER                  = 4
	IT_SERIAL                 = 8
	IT_JOYPAD                 = 16
)

func IntHandle(ctx *CpuContext, address uint16) {
	StackPush16(ctx.Regs.Pc)
	ctx.Regs.Pc = address
}

// is this correct?
func IntCheck(ctx *CpuContext, address uint16, it InterruptType) bool {

	if (ctx.IntFlags&byte(it)) == 1 && (ctx.IERegister&byte(it) == 1) {
		IntHandle(ctx, address)
		ctx.IntFlags &= ^byte(it)
		ctx.Halted = false
		ctx.IntMasterEnabled = false

		return true
	}

	return false
}

func CpuHandleInterrupts(ctx *CpuContext) {
	if IntCheck(ctx, 0x40, IT_VBLANK) {

	} else if IntCheck(ctx, 0x48, IT_LCD_STAT) {

	} else if IntCheck(ctx, 0x50, IT_TIMER) {

	} else if IntCheck(ctx, 0x58, IT_SERIAL) {

	} else if IntCheck(ctx, 0x60, IT_JOYPAD) {

	}
}
