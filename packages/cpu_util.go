package gameboypackage

func CpuFlagZ() bool {
	return Bit(CpuCtx.Regs.f, 7)
}

func CpuFlagC() bool {
	return Bit(CpuCtx.Regs.f, 4)
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
		return Reverse(uint16(CpuCtx.Regs.a))
	case RT_BC:
		return Reverse(uint16(CpuCtx.Regs.b))
	case RT_DE:
		return Reverse(uint16(CpuCtx.Regs.d))
	case RT_HL:
		return Reverse(uint16(CpuCtx.Regs.h))

	case RT_PC:
		return CpuCtx.Regs.pc
	case RT_SP:
		return CpuCtx.Regs.sp
	default:
		return 0
	}
}

//could be problem with cast here
func CpuSetReg(regType regTypes, val uint16) {
	switch regType {
	case RT_A:
		CpuCtx.Regs.a = byte(val & 0xFF)
		break
	case RT_F:
		CpuCtx.Regs.f = byte(val & 0xFF)
		break
	case RT_B:
		CpuCtx.Regs.b = byte(val & 0xFF)
		break
	case RT_C:
		CpuCtx.Regs.c = byte(val & 0xFF)
		break
	case RT_D:
		CpuCtx.Regs.d = byte(val & 0xFF)
		break
	case RT_E:
		CpuCtx.Regs.e = byte(val & 0xFF)
		break
	case RT_H:
		CpuCtx.Regs.h = byte(val & 0xFF)
		break
	case RT_L:
		CpuCtx.Regs.l = byte(val & 0xFF)
		break

	case RT_AF:
		CpuCtx.Regs.a = byte(Reverse(val))
		break
	case RT_BC:
		CpuCtx.Regs.b = byte(Reverse(val))
		break
	case RT_DE:
		CpuCtx.Regs.d = byte(Reverse(val))
		break
	case RT_HL:
		CpuCtx.Regs.h = byte(Reverse(val))
		break

	case RT_PC:
		CpuCtx.Regs.pc = val
		break
	case RT_SP:
		CpuCtx.Regs.sp = val
		break
	case RT_NONE:
		break
	}
}
