package cpu

type InterruptType byte

const (
	IT_VBLANK   InterruptType = 1 << iota // 0x01
	IT_LCD_STAT                           // 0x02
	IT_TIMER                              // 0x04
	IT_SERIAL                             // 0x08
	IT_JOYPAD                             // 0x10
)

func IntHandle(ctx *CpuContext, address uint16, it InterruptType) {
	// Disable interrupts (IME = 0)
	ctx.IntMasterEnabled = false

	// Clear the corresponding interrupt flag
	ctx.IntFlags &= ^byte(it)

	// Push current PC to stack
	StackPush16(ctx.Regs.Pc)

	// Jump to interrupt vector
	ctx.Regs.Pc = address
}

func IntCheck(ctx *CpuContext, address uint16, it InterruptType) bool {
	if !ctx.IntMasterEnabled {
		return false
	}

	ieRegister := ctx.memoryBus.BusRead(0xFFFF) // Read IE from bus, not cached copy

	ifFlag := (ctx.IntFlags & byte(it)) != 0
	ieFlag := (ieRegister & byte(it)) != 0

	if ifFlag && ieFlag {
		IntHandle(ctx, address, it)
		ctx.Halted = false
		Cm.IncreaseCycle(2) // Interrupt handling takes additional cycles
		return true
	}
	return false
}

func CpuHandleInterrupts(ctx *CpuContext) {
	if IntCheck(ctx, 0x40, IT_VBLANK) {
		return
	} else if IntCheck(ctx, 0x48, IT_LCD_STAT) {
		return
	} else if IntCheck(ctx, 0x50, IT_TIMER) {
		return
	} else if IntCheck(ctx, 0x58, IT_SERIAL) {
		return
	} else if IntCheck(ctx, 0x60, IT_JOYPAD) {
		return
	}
}
