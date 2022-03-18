package cpu

import (
	"log"

	"pajalic.go.emulator/packages/common"
	"pajalic.go.emulator/packages/emulator"
)

func procNone(ctx *CpuContext) {
	log.Fatal("Invalid Instruction!")
}

func procLd(ctx *CpuContext) {
	if ctx.DestIsMem {
		//LD (BC), A for instance...

		if ctx.currentInst.Reg2 >= RT_AF {
			//if 16 bit register...
			emulator.EmuCycles(1)
			BusWrite16(ctx.MemDest, ctx.FetchedData)
		} else {
			BusWrite(ctx.MemDest, byte(ctx.FetchedData))
		}

		return
	}

	if ctx.currentInst.Mode == AM_HL_SPR {
		var hflag = (CpuRegRead(ctx.currentInst.Reg2)&0xF)+
			(ctx.FetchedData&0xF) >= 0x10

		var cflag = (CpuRegRead(ctx.currentInst.Reg2)&0xFF)+
			(ctx.FetchedData&0xFF) >= 0x100

		CpuSetFlags(*ctx, nil, nil, &hflag, &cflag)
		CpuSetReg(ctx.currentInst.Reg1,
			CpuRegRead(ctx.currentInst.Reg2)+ctx.FetchedData)

		return
	}
	//probleem när ctx.FetchedData är 143? Efter CALL?
	CpuSetReg(ctx.currentInst.Reg1, ctx.FetchedData)
}

func procNop(cpucontext *CpuContext) {

}

func procDi(ctx *CpuContext) {
	ctx.IntMasterEnabled = false
}

func procPop(ctx *CpuContext) {
	var lo = uint16(StackPop())
	emulator.EmuCycles(1)
	var hi = uint16(StackPop())
	emulator.EmuCycles(1)
	var n = (hi << 8) | lo
	CpuSetReg(ctx.currentInst.Reg1, n)

	if ctx.currentInst.Reg1 == RT_AF {
		CpuSetReg(ctx.currentInst.Reg1, n&0xFFF0)
	}
}

func procPush(ctx *CpuContext) {
	var hi = (CpuRegRead(ctx.currentInst.Reg1) >> 8) & 0xFF
	emulator.EmuCycles(1)
	StackPush(byte(hi))
	var lo = (CpuRegRead(ctx.currentInst.Reg2)) & 0xFF
	emulator.EmuCycles(1)
	StackPush(byte(lo))
	emulator.EmuCycles(1)
}

func goToAddr(ctx *CpuContext, addr uint16, pushpc bool) {
	if CheckCondition(ctx) {
		if pushpc {
			emulator.EmuCycles(2)
			StackPush16(ctx.Regs.pc)
		}
		ctx.Regs.pc = addr
		emulator.EmuCycles(1)
	}
}

func procJp(ctx *CpuContext) {
	goToAddr(ctx, ctx.FetchedData, false)
}

//Jump relative
func procJr(ctx *CpuContext) {
	var rel = byte(ctx.FetchedData & 0xFF) //casting cause it might be negative
	var addr = ctx.Regs.pc + uint16(rel)
	goToAddr(ctx, addr, false)
}

func procCall(ctx *CpuContext) {
	goToAddr(ctx, ctx.FetchedData, true)
}

//0000002D
func procRet(ctx *CpuContext) {
	if ctx.currentInst.Condition != CT_NONE {
		emulator.EmuCycles(1)
	}

	if CheckCondition(ctx) {
		var lo = uint16(StackPop())
		emulator.EmuCycles(1)

		var hi = uint16(StackPop())
		emulator.EmuCycles(1)

		var n = (hi << 8) | lo
		ctx.Regs.pc = n

		emulator.EmuCycles(1)
	}
}

func procRst(ctx *CpuContext) {
	goToAddr(ctx, uint16(ctx.currentInst.Param), true)
}

func procReti(ctx *CpuContext) {
	ctx.IntMasterEnabled = true
	procRet(ctx)
}

func procLdh(ctx *CpuContext) {
	//Ensure this is proper
	a := uint16(BusRead(0xFF00 | uint16(ctx.FetchedData)))

	if ctx.currentInst.Reg1 == RT_A {
		CpuSetReg(ctx.currentInst.Reg1, a)
	} else {
		BusWrite(ctx.MemDest, ctx.Regs.a)
	}
	emulator.EmuCycles(1)
}

func procXor(ctx *CpuContext) {
	ctx.Regs.a ^= byte(ctx.FetchedData & 0xFF)
	var zflag = false
	if ctx.Regs.a == 0 {
		zflag = true
	}
	CpuSetFlags(*ctx, &zflag, nil, nil, nil)

}

func procInc(ctx *CpuContext) {
	var val = CpuRegRead(ctx.currentInst.Reg1) + 1

	if is16bit(ctx.currentInst.Reg1) {
		emulator.EmuCycles(1)
	}

	if ctx.currentInst.Reg1 == RT_HL && ctx.currentInst.Mode == AM_MR {
		val = uint16(BusRead(CpuRegRead(RT_HL)) + 1)
		val &= 0xFF
		BusWrite(CpuRegRead(RT_HL), byte(val))
	} else {
		CpuSetReg(ctx.currentInst.Reg1, val)
		val = CpuRegRead(ctx.currentInst.Reg1)
	}

	if (ctx.CurOpCode & 0x03) == 0x03 {
		return
	}
	n := false
	z := val == 0
	h := (val & 0x0F) == 0
	CpuSetFlags(*ctx, &z, &n, &h, nil)
}

func procDec(ctx *CpuContext) {
	var val = CpuRegRead(ctx.currentInst.Reg1) - 1

	if is16bit(ctx.currentInst.Reg1) {
		emulator.EmuCycles(1)
	}

	if ctx.currentInst.Reg1 == RT_HL && ctx.currentInst.Mode == AM_MR {
		val = uint16(BusRead(CpuRegRead(RT_HL)) - 1)
		BusWrite(CpuRegRead(RT_HL), byte(val))
	} else {
		CpuSetReg(ctx.currentInst.Reg1, val)
		val = CpuRegRead(ctx.currentInst.Reg1)
	}

	if (ctx.CurOpCode & 0x0B) == 0x0B {
		return
	}
	n := true
	z := val == 0
	h := (val & 0x0F) == 0x0F
	CpuSetFlags(*ctx, &z, &n, &h, nil)
}

func procSub(ctx *CpuContext) {
	var val = CpuRegRead(ctx.currentInst.Reg1) - ctx.FetchedData

	var z = val == 0
	var h = (int32(CpuRegRead(ctx.currentInst.Reg1&0xF)) - int32(ctx.FetchedData&0xF)) < 0
	var c = (int32(CpuRegRead(ctx.currentInst.Reg1)) - int32(ctx.FetchedData)) < 0

	CpuSetReg(ctx.currentInst.Reg1, val)

	n := true
	CpuSetFlags(*ctx, &z, &n, &h, &c)
}

func procSbc(ctx *CpuContext) {
	var c uint16 = 0
	CpuFlagC()
	if CpuFlagC() {
		c = 1
	}
	var val = byte(ctx.FetchedData + c)

	var z = CpuRegRead(ctx.currentInst.Reg1)-uint16(val) == 0

	var h = (int32(CpuRegRead(ctx.currentInst.Reg1)&0xF) - int32(ctx.FetchedData&0xF) - int32(c)) < 0

	var cf = (int32(CpuRegRead(ctx.currentInst.Reg1)) - int32(ctx.FetchedData) - int32(c)) < 0

	CpuSetReg(ctx.currentInst.Reg1, CpuRegRead(ctx.currentInst.Reg1)-uint16(val))

	n := true
	CpuSetFlags(*ctx, &z, &n, &h, &cf)
}

func procAdc(ctx *CpuContext) {
	var u = ctx.FetchedData
	var a = uint16(ctx.Regs.a)
	var c uint16 = 0
	if CpuFlagC() {
		c = 1
	}

	ctx.Regs.a = byte((a + u + c) & 0xFF)

	zf := ctx.Regs.a == 0
	hf := (a&0xF)+(u&0xF)+c > 0xF
	cf := (a + u + c) > 0xFF

	n := false
	CpuSetFlags(*ctx, &zf, &n, &hf, &cf)
}

//Bool to Int problematik som sker i C men ej i GOlang
func procAdd(ctx *CpuContext) {
	var val = uint32(CpuRegRead(ctx.currentInst.Reg1) + ctx.FetchedData)

	var is16bit = is16bit(ctx.currentInst.Reg1)

	if is16bit {
		emulator.EmuCycles(1)
	}

	if ctx.currentInst.Reg1 == RT_SP {
		//prevent overflow
		ctxFetchedByte := byte(ctx.FetchedData)
		val = uint32(CpuRegRead(ctx.currentInst.Reg1) + uint16(ctxFetchedByte))
	}

	var z = (val & 0xFF) == 0
	var h = (CpuRegRead(ctx.currentInst.Reg1)&0xF)+(ctx.FetchedData&0xF) >= 0x10
	var c = (int)(CpuRegRead(ctx.currentInst.Reg1)&0xFF)+(int)(ctx.FetchedData&0xFF) >= 0x100

	zptr := &z
	hptr := &h
	cptr := &c

	if is16bit {
		zptr = nil
		*hptr = (CpuRegRead(ctx.currentInst.Reg1)&0xFFF)+(ctx.FetchedData&0xFFF) >= 0x1000
		n := (uint32(CpuRegRead(ctx.currentInst.Reg1))) + uint32(ctx.FetchedData)
		*cptr = n >= 0x10000
	}

	if ctx.currentInst.Reg1 == RT_SP {
		zptr = nil
		h = (CpuRegRead(ctx.currentInst.Reg1)&0xF)+(ctx.FetchedData&0xF) >= 0x10
		c = (int32)(CpuRegRead(ctx.currentInst.Reg1)&0xFF)+(int32)(ctx.FetchedData&0xFF) > 0x100
	}

	n := false

	CpuSetReg(ctx.currentInst.Reg1, uint16(val&0xFFFF))
	CpuSetFlags(*ctx, zptr, &n, hptr, cptr)
}

func is16bit(rt regTypes) bool {
	return rt >= RT_AF
}
func CheckCondition(ctx *CpuContext) bool {
	z := CpuFlagZ()
	c := CpuFlagC()

	switch CpuCtx.currentInst.Condition {
	case CT_NONE:
		return true
	case CT_C:
		return c
	case CT_NC:
		return !c
	case CT_Z:
		return z
	case CT_NZ:
		return !z

	}
	return false

}

//Function pointer MAP
type InProc func(ctx *CpuContext)

var processors = make(map[InType]InProc)

func InitProcessors() {
	processors[IN_NONE] = procNone
	processors[IN_NOP] = procNop
	processors[IN_LD] = procLd
	processors[IN_JP] = procJp
	processors[IN_JR] = procJr
	processors[IN_CALL] = procCall
	processors[IN_XOR] = procXor
	processors[IN_POP] = procPop
	processors[IN_PUSH] = procPush
	processors[IN_LDH] = procLdh
	processors[IN_RET] = procRet
	processors[IN_RETI] = procReti
	processors[IN_DI] = procDi
	processors[IN_RST] = procRst
	processors[IN_ADD] = procAdd
	processors[IN_ADC] = procAdc
	processors[IN_INC] = procInc
	processors[IN_DEC] = procDec
	processors[IN_SUB] = procSub
	processors[IN_SBC] = procSbc
}

//fix this to return properly
func InstGetProccessor(intype InType) InProc {

	return processors[intype]

}

func CpuSetFlags(ctx CpuContext, z *bool, n *bool, h *bool, c *bool) {
	if z != nil {
		common.BitSet(&ctx.Regs.f, 7, z)
	}

	if n != nil {
		common.BitSet(&ctx.Regs.f, 6, n)
	}

	if h != nil {
		common.BitSet(&ctx.Regs.f, 5, h)
	}

	if c != nil {
		common.BitSet(&ctx.Regs.f, 4, c)
	}
}
