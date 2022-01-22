package gameboypackage

//reverse byte how is this done, can it be easier
func reverse(n uint16) uint16 {
	return ((n & 0xFF00) >> 8) | ((n & 0x00FF) << 8)
}

func CpuRegRead(regType regTypes) uint16 {
	switch regType {
	case RT_A:
		return uint16(CpuCtx.Regs.a)
	case RT_F:
		return uint16(CpuCtx.Regs.f)
	case RT_B:
		return uint16(CpuCtx.Regs.b)
	case RT_C:
		return uint16(CpuCtx.Regs.c)
	case RT_D:
		return uint16(CpuCtx.Regs.d)
	case RT_E:
		return uint16(CpuCtx.Regs.e)
	case RT_H:
		return uint16(CpuCtx.Regs.h)
	case RT_L:
		return uint16(CpuCtx.Regs.l)
	//Pointer magic here?
	case RT_AF:
		return reverse(uint16(CpuCtx.Regs.a))
	case RT_BC:
		return reverse(uint16(CpuCtx.Regs.b))
	case RT_DE:
		return reverse(uint16(CpuCtx.Regs.d))
	case RT_HL:
		return reverse(uint16(CpuCtx.Regs.h))

	case RT_PC:
		return CpuCtx.Regs.pc
	case RT_SP:
		return CpuCtx.Regs.sp
	default:
		return 0
	}
}
