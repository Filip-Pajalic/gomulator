package cpu

type InterruptType byte

const (
	IT_VBLANK   InterruptType = 1 << iota // 0x01
	IT_LCD_STAT                           // 0x02
	IT_TIMER                              // 0x04
	IT_SERIAL                             // 0x08
	IT_JOYPAD                             // 0x10
)

func IntHandle(ctx *CpuContext, address uint16) {
	StackPush16(ctx.Regs.Pc)
	ctx.Regs.Pc = address
}

func IntCheck(ctx *CpuContext, address uint16, it InterruptType) bool {
	if (ctx.IntFlags&byte(it)) != 0 && (ctx.iERegister&byte(it)) != 0 {
		IntHandle(ctx, address)
		ctx.IntFlags &= ^byte(it) // Clear the interrupt flag
		ctx.Halted = false
		ctx.IntMasterEnabled = false
		return true
	}
	return false
}

func CpuHandleInterrupts(ctx *CpuContext) {
	if IntCheck(ctx, 0x40, IT_VBLANK) {
		// VBLANK interrupt handled
	} else if IntCheck(ctx, 0x48, IT_LCD_STAT) {
		// LCD STAT interrupt handled
	} else if IntCheck(ctx, 0x50, IT_TIMER) {
		// TIMER interrupt handled
	} else if IntCheck(ctx, 0x58, IT_SERIAL) {
		// SERIAL interrupt handled
	} else if IntCheck(ctx, 0x60, IT_JOYPAD) {
		// JOYPAD interrupt handled
	}
}
