package cpu

import (
	"pajalic.go.emulator/packages/common"
	log "pajalic.go.emulator/packages/logger"
)

func CpuFlagZ() bool {
	return common.Bit(CpuCtx.Regs.f, 7)
}

func CpuFlagN() bool {
	return common.Bit(CpuCtx.Regs.f, 6)
}

func CpuFlagH() bool {
	return common.Bit(CpuCtx.Regs.f, 5)
}

func CpuFlagC() bool {
	return common.Bit(CpuCtx.Regs.f, 4)
}

//Broken here
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
	case RT_AF:
		return common.Reverse(uint16(CpuCtx.Regs.f)<<8 | uint16(CpuCtx.Regs.a))
	case RT_BC:
		return common.Reverse(uint16(CpuCtx.Regs.c)<<8 | uint16(CpuCtx.Regs.b))
	case RT_DE:
		return common.Reverse(uint16(CpuCtx.Regs.e)<<8 | uint16(CpuCtx.Regs.d))
	case RT_HL:
		return common.Reverse(uint16(CpuCtx.Regs.l)<<8 | uint16(CpuCtx.Regs.h))
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
		result := common.Reverse(val)
		CpuCtx.Regs.f = byte((result >> 8) & 0xFF)
		CpuCtx.Regs.a = byte((result) & 0xFF)
		break
	case RT_BC:
		result := common.Reverse(val)
		CpuCtx.Regs.c = byte((result >> 8) & 0xFF)
		CpuCtx.Regs.b = byte((result) & 0xFF)
		break
	case RT_DE:
		result := common.Reverse(val)
		CpuCtx.Regs.e = byte((result >> 8) & 0xFF)
		CpuCtx.Regs.d = byte((result) & 0xFF)
		break
	case RT_HL:
		result := common.Reverse(val)
		CpuCtx.Regs.l = byte((result >> 8) & 0xFF)
		CpuCtx.Regs.h = byte((result) & 0xFF)
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

func CpuRegRead8(rt regTypes) byte {
	switch rt {
	case RT_A:
		return CpuCtx.Regs.a
	case RT_F:
		return CpuCtx.Regs.f
	case RT_B:
		return CpuCtx.Regs.b
	case RT_C:
		return CpuCtx.Regs.c
	case RT_D:
		return CpuCtx.Regs.d
	case RT_E:
		return CpuCtx.Regs.e
	case RT_H:
		return CpuCtx.Regs.h
	case RT_L:
		return CpuCtx.Regs.l
	case RT_HL:
		{
			return BusRead(CpuRegRead(RT_HL))
		}
	default:
		log.Fatal("**ERR INVALID REG8: %d\n", rt)

	}
	return 0
}

func CpuSetReg8(rt regTypes, val byte) {
	switch rt {
	case RT_A:
		CpuCtx.Regs.a = val & 0xFF
		break
	case RT_F:
		CpuCtx.Regs.f = val & 0xFF
		break
	case RT_B:
		CpuCtx.Regs.b = val & 0xFF
		break
	case RT_C:
		CpuCtx.Regs.c = val & 0xFF
		break
	case RT_D:
		CpuCtx.Regs.d = val & 0xFF
		break
	case RT_E:
		CpuCtx.Regs.e = val & 0xFF
		break
	case RT_H:
		CpuCtx.Regs.h = val & 0xFF
		break
	case RT_L:
		CpuCtx.Regs.l = val & 0xFF
		break
	case RT_HL:
		BusWrite(CpuRegRead(RT_HL), val)
		break
	default:
		log.Fatal("**ERR INVALID REG8: %d\n", rt)
	}
}

func CpuGetRegs() *CpuRegisters {
	return &CpuCtx.Regs
}

func CpuGetIntFlags() byte {
	return CpuCtx.IntFlags
}

func CpuSetIntFlags(value byte) {
	CpuCtx.IntFlags = value
}
