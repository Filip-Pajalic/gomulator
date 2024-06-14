package cpu

import (
	"log"
)

func procNone(ctx *CpuContext) {
	log.Fatal("Invalid Instruction!")
}

var rtLookup = []regTypes{
	RT_B,
	RT_C,
	RT_D,
	RT_E,
	RT_H,
	RT_L,
	RT_HL,
	RT_A,
}

// probably wrong , probably needs to have a byte assigned to each rT
func decodeReg(reg byte) regTypes {
	if reg > 0b111 {
		return RT_NONE
	}

	return rtLookup[reg]
}

func procNop(cpucontext *CpuContext) {

}

func procLd(ctx *CpuContext) {
	if ctx.DestIsMem {
		//LD (BC), A for instance...

		if is16bit(ctx.currentInst.Reg2) {
			//if 16 bit register...
			EmuCycles(1)
			BusWrite16(ctx.MemDest, ctx.FetchedData)
		} else {
			BusWrite(ctx.MemDest, byte(ctx.FetchedData))
		}
		EmuCycles(1)
		return
	}

	if ctx.currentInst.Mode == AM_HL_SPR {

		var hflag = ((CpuRegRead(ctx.currentInst.Reg2) & 0x0F) + (ctx.FetchedData & 0x0F)) >= 0x10

		// Calculate cflag
		var cflag = ((CpuRegRead(ctx.currentInst.Reg2) & 0xFF) + (uint16(ctx.FetchedData) & 0xFF)) >= 0x100

		zflag := false
		nflag := false

		// Set flags
		CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)

		// Cast fetched data to int8
		signedFetchedData := int8(ctx.FetchedData)

		// Set register value
		CpuSetReg(ctx.currentInst.Reg1, CpuRegRead(ctx.currentInst.Reg2)+uint16(signedFetchedData))

		return
	}
	//probleem när ctx.FetchedData är 143? Efter CALL?
	CpuSetReg(ctx.currentInst.Reg1, ctx.FetchedData)
}

func procCb(ctx *CpuContext) {
	var op = byte(ctx.FetchedData)
	var reg = decodeReg(op & 0b111)
	var bit = (op >> 3) & 0b111
	var bit_op = (op >> 6) & 0b11
	var regval = CpuRegRead8(reg)
	EmuCycles(1)

	if reg == RT_HL {
		EmuCycles(2)
	}

	switch bit_op {
	case 0:
		// BIT
		zflag := (regval & (1 << bit)) == 0
		nflag := false
		hflag := true
		CpuSetFlags(ctx, &zflag, &nflag, &hflag, nil)
		return

	case 1:
		// RES
		regval &= ^(1 << bit)
		CpuSetReg8(reg, regval)
		return

	case 2:
		// SET
		regval |= 1 << bit
		CpuSetReg8(reg, regval)
		return

	case 3:
		// SWAP
		regval = ((regval & 0xF0) >> 4) | ((regval & 0x0F) << 4)
		zflag := regval == 0
		nflag := false
		hflag := false
		cflag := false
		CpuSetReg8(reg, regval)
		CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
		return
	}

	flagC := CpuFlagC()

	switch bit {
	case 0:
		// RLC
		setC := false
		result := (regval << 1) | (regval >> 7)
		zflag := result == 0
		nflag := false
		hflag := false
		if (regval & 0x80) != 0 {
			setC = true
		}
		CpuSetReg8(reg, result)
		CpuSetFlags(ctx, &zflag, &nflag, &hflag, &setC)
		return

	case 1:
		// RRC
		setC := false
		result := (regval >> 1) | ((regval & 1) << 7)
		zflag := result == 0
		nflag := false
		hflag := false
		if (regval & 1) != 0 {
			setC = true
		}
		CpuSetReg8(reg, result)
		CpuSetFlags(ctx, &zflag, &nflag, &hflag, &setC)
		return

	case 2:
		// RL
		setC := false
		result := (regval << 1) | boolToByte(flagC)
		zflag := result == 0
		nflag := false
		hflag := false
		if (regval & 0x80) != 0 {
			setC = true
		}
		CpuSetReg8(reg, result)
		CpuSetFlags(ctx, &zflag, &nflag, &hflag, &setC)
		return

	case 3:
		// RR
		setC := false
		result := (regval >> 1) | (boolToByte(flagC) << 7)
		zflag := result == 0
		nflag := false
		hflag := false
		if (regval & 1) != 0 {
			setC = true
		}
		CpuSetReg8(reg, result)
		CpuSetFlags(ctx, &zflag, &nflag, &hflag, &setC)
		return

	case 4:
		// SLA
		setC := false
		result := regval << 1
		zflag := result == 0
		nflag := false
		hflag := false
		if (regval & 0x80) != 0 {
			setC = true
		}
		CpuSetReg8(reg, result)
		CpuSetFlags(ctx, &zflag, &nflag, &hflag, &setC)
		return

	case 5:
		// SRA
		setC := false
		result := ((regval) >> 1) | (regval & 0x80)
		zflag := result == 0
		nflag := false
		hflag := false
		if (regval & 1) != 0 {
			setC = true
		}
		CpuSetReg8(reg, result)
		CpuSetFlags(ctx, &zflag, &nflag, &hflag, &setC)
		return

	case 6:
		// SWAP
		result := (regval >> 4) | (regval << 4)
		zflag := result == 0
		nflag := false
		hflag := false
		cflag := false
		CpuSetReg8(reg, result)
		CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
		return

	case 7:
		// SRL
		setC := false
		result := regval >> 1
		zflag := result == 0
		nflag := false
		hflag := false
		if (regval & 1) != 0 {
			setC = true
		}
		CpuSetReg8(reg, result)
		CpuSetFlags(ctx, &zflag, &nflag, &hflag, &setC)
		return
	}

	log.Fatalf("ERROR: INVALID CB: %02X", op)
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

// Could be a problem with casting here
func procAnd(ctx *CpuContext) {
	ctx.Regs.A &= byte(ctx.FetchedData)
	var zflag = ctx.Regs.A == 0
	var nflag = false
	var hflag = true
	var cflag = false

	CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
}

func procRlca(ctx *CpuContext) {

	var u = ctx.Regs.A

	z := false
	n := false
	h := false
	c := (u>>7)&1 == 1
	u = (u << 1)
	if c {
		u |= 1
	}

	ctx.Regs.A = u

	CpuSetFlags(ctx, &z, &n, &h, &c)

}

func procRrca(ctx *CpuContext) {
	var b = ctx.Regs.A & 1
	ctx.Regs.A >>= 1
	ctx.Regs.A |= b << 7
	var zflag = false
	var nflag = false
	var hflag = false
	var cflag = b == 1

	CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)

}

func procRla(ctx *CpuContext) {
	var u = ctx.Regs.A

	var zflag = false
	var nflag = false
	var hflag = false
	var cflag = (u>>7)&1 == 1

	var cf byte = 0
	if CpuFlagC() {
		cf = 1
	}

	ctx.Regs.A = (u << 1) | cf

	CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)

}
func procRra(ctx *CpuContext) {
	var new_c = ctx.Regs.A & 1
	var zflag = false
	var nflag = false
	var hflag = false
	var cflag = new_c == 1

	ctx.Regs.A >>= 1
	var carry byte = 0
	if CpuFlagC() {
		carry = 1
	}
	ctx.Regs.A |= carry << 7

	CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)

}

func procXor(ctx *CpuContext) {
	ctx.Regs.A ^= byte(ctx.FetchedData & 0xFF)
	var zflag = ctx.Regs.A == 0
	var nflag = false
	var hflag = false
	var cflag = false

	CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)

}

func procOr(ctx *CpuContext) {
	ctx.Regs.A |= byte(ctx.FetchedData & 0xFF)
	zflag := ctx.Regs.A == 0
	nflag := false
	hflag := false
	cflag := false

	CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
}

func procCp(ctx *CpuContext) {

	//ctx reg a wrong is 1 too much how
	n := int(ctx.Regs.A) - int(ctx.FetchedData)

	zflag := n == 0
	nflag := true
	hflag := (((int)(ctx.Regs.A) & 0x0F) - ((int)(ctx.FetchedData) & 0x0F)) < 0
	cflag := n < 0

	CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
}

func procDi(ctx *CpuContext) {
	ctx.IntMasterEnabled = false
}

func procEi(ctx *CpuContext) {
	ctx.enablingIme = true
}

func procPop(ctx *CpuContext) {
	var lo = uint16(StackPop())
	EmuCycles(1)
	var hi = uint16(StackPop())
	EmuCycles(1)
	var n = (hi << 8) | lo
	CpuSetReg(ctx.currentInst.Reg1, n)

	if ctx.currentInst.Reg1 == RT_AF {
		CpuSetReg(ctx.currentInst.Reg1, n&0xFFF0)
	}
}

func procPush(ctx *CpuContext) {
	var hi = (CpuRegRead(ctx.currentInst.Reg1) >> 8) & 0xFF
	EmuCycles(1)
	StackPush(byte(hi))
	var lo = (CpuRegRead(ctx.currentInst.Reg1)) & 0xFF
	EmuCycles(1)
	StackPush(byte(lo))
	EmuCycles(1)
}

func goToAddr(ctx *CpuContext, addr uint16, pushpc bool) {
	if CheckCondition(ctx) {
		if pushpc {
			EmuCycles(2)
			StackPush16(ctx.Regs.Pc)
		}
		ctx.Regs.Pc = addr
		EmuCycles(1)
	}
}

func procJp(ctx *CpuContext) {
	goToAddr(ctx, ctx.FetchedData, false)
}

// Jump relative
func procJr(ctx *CpuContext) {
	var rel = int8(ctx.FetchedData & 0xFF) //casting cause it might be negative
	var addr = ctx.Regs.Pc + uint16(rel)
	goToAddr(ctx, addr, false)
}

func procCall(ctx *CpuContext) {
	goToAddr(ctx, ctx.FetchedData, true)
}

func procRet(ctx *CpuContext) {
	if ctx.currentInst.Condition != CT_NONE {
		EmuCycles(1)
	}

	if CheckCondition(ctx) {
		var lo = uint16(StackPop())
		EmuCycles(1)

		var hi = uint16(StackPop())
		EmuCycles(1)

		var n = (hi << 8) | lo
		ctx.Regs.Pc = n

		EmuCycles(1)
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
	//is CpuSetReg correct here
	if ctx.currentInst.Reg1 == RT_A {
		CpuSetReg(ctx.currentInst.Reg1, uint16(BusRead(0xFF00|ctx.FetchedData)))
	} else {
		BusWrite(ctx.MemDest, ctx.Regs.A)
	}
	EmuCycles(1)
}

func procInc(ctx *CpuContext) {
	var val = CpuRegRead(ctx.currentInst.Reg1) + 1

	if is16bit(ctx.currentInst.Reg1) {
		EmuCycles(1)
	}

	if ctx.currentInst.Reg1 == RT_HL && ctx.currentInst.Mode == AM_MR {
		val = uint16(BusRead(CpuRegRead(RT_HL))) + 1
		val &= 0xFF
		BusWrite(CpuRegRead(RT_HL), byte(val))
	} else {
		CpuSetReg(ctx.currentInst.Reg1, val)
		val = CpuRegRead(ctx.currentInst.Reg1)
	}

	if (ctx.CurOpCode & 0x03) == 0x03 {
		return
	}
	z := val == 0
	n := false
	h := (val & 0x0F) == 0
	CpuSetFlags(ctx, &z, &n, &h, nil)
}

func procDec(ctx *CpuContext) {
	var val = CpuRegRead(ctx.currentInst.Reg1) - 1

	if is16bit(ctx.currentInst.Reg1) {
		EmuCycles(1)
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
	CpuSetFlags(ctx, &z, &n, &h, nil)
}

func procSub(ctx *CpuContext) {

	var val = CpuRegRead(ctx.currentInst.Reg1) - ctx.FetchedData

	var z = val == 0
	var h = (int32(CpuRegRead(ctx.currentInst.Reg1)&0xF) - int32(ctx.FetchedData&0xF)) < 0
	var c = (int32(CpuRegRead(ctx.currentInst.Reg1)) - int32(ctx.FetchedData)) < 0

	CpuSetReg(ctx.currentInst.Reg1, val)

	n := true
	CpuSetFlags(ctx, &z, &n, &h, &c)
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
	CpuSetFlags(ctx, &z, &n, &h, &cf)
}

func procAdc(ctx *CpuContext) {

	var u = ctx.FetchedData
	var a = uint16(ctx.Regs.A)
	var c uint16 = 0
	if CpuFlagC() {
		c = 1
	}

	ctx.Regs.A = byte((a + u + c) & 0xFF)

	z := ctx.Regs.A == 0
	n := false
	h := (a&0xF)+(u&0xF)+c > 0xF
	cb := a+u+c > 0xFF

	CpuSetFlags(ctx, &z, &n, &h, &cb)
}

// Bool to Int problematik som sker i C men ej i GOlang
func procAdd(ctx *CpuContext) {

	var val = uint32(CpuRegRead(ctx.currentInst.Reg1) + ctx.FetchedData)

	var is16bit = is16bit(ctx.currentInst.Reg1)

	if is16bit {
		EmuCycles(1)
	}

	if ctx.currentInst.Reg1 == RT_SP {
		val = uint32(CpuRegRead(ctx.currentInst.Reg1) + uint16(byte(ctx.FetchedData)))
	}

	z := (val & 0xFF) == 0

	h := (CpuRegRead(ctx.currentInst.Reg1)&0xF)+(ctx.FetchedData&0xF) >= 0x10
	c := (int)(CpuRegRead(ctx.currentInst.Reg1)&0xFF)+(int)(ctx.FetchedData&0xFF) >= 0x100
	n := false

	if is16bit {
		h = (CpuRegRead(ctx.currentInst.Reg1)&0xFFF)+(ctx.FetchedData&0xFFF) >= 0x1000
		c = (uint32(CpuRegRead(ctx.currentInst.Reg1)))+uint32(ctx.FetchedData) >= 0x10000

	}

	if ctx.currentInst.Reg1 == RT_SP {
		z = false
		h = (CpuRegRead(ctx.currentInst.Reg1)&0xF)+(ctx.FetchedData&0xF) >= 0x10
		c = (int32)(CpuRegRead(ctx.currentInst.Reg1)&0xFF)+(int32)(ctx.FetchedData&0xFF) >= 0x100

		CpuSetReg(ctx.currentInst.Reg1, uint16(val&0xFFFF))
		CpuSetFlags(ctx, &z, &n, &h, &c)
		return
	}

	if is16bit {
		CpuSetReg(ctx.currentInst.Reg1, uint16(val&0xFFFF))
		CpuSetFlags(ctx, nil, &n, &h, &c)
		return
	}

	CpuSetReg(ctx.currentInst.Reg1, uint16(val&0xFFFF))
	CpuSetFlags(ctx, &z, &n, &h, &c)
}

func procStop(ctx *CpuContext) {
	log.Fatal("STOPPING!")
}

func procDaa(ctx *CpuContext) {
	var u byte = 0
	var fc = 0
	//   if (CPU_FLAG_H || (!CPU_FLAG_N && (ctx->regs.a & 0xF) > 9)) {
	//        u = 6;
	//    }
	if CpuFlagH() || (!CpuFlagN() && (ctx.Regs.A&0xF) > 9) {
		u = 6
	}
	//    if (CPU_FLAG_C || (!CPU_FLAG_N && ctx->regs.a > 0x99)) {
	//        u |= 0x60;
	//        fc = 1;
	//    }
	if CpuFlagC() || (!CpuFlagN() && ctx.Regs.A > 0x99) {
		u |= 0x60
		fc = 1
	}

	//ctx->regs.a += CPU_FLAG_N ? -u : u;
	if CpuFlagN() {
		ctx.Regs.A += -u
	} else {
		ctx.Regs.A += u
	}

	//cpu_set_flags(ctx, ctx->regs.a == 0, -1, 0, fc);

	z := ctx.Regs.A == 0
	h := false
	c := fc == 1

	if ctx.Regs.A == 0 {
		z = true
	}

	CpuSetFlags(ctx, &z, nil, &h, &c)

}

func procCpl(ctx *CpuContext) {
	ctx.Regs.A = ^ctx.Regs.A
	var nflag = true
	var hflag = true

	CpuSetFlags(ctx, nil, &nflag, &hflag, nil)
}

func procScf(ctx *CpuContext) {
	var nflag = false
	var hflag = false
	var cflag = true

	CpuSetFlags(ctx, nil, &nflag, &hflag, &cflag)
}

func procCcf(ctx *CpuContext) {

	var nflag = false
	var hflag = false
	var cflag = false
	var cpuflagbit byte = 0
	if CpuFlagC() {
		cpuflagbit = 1
	}
	if (cpuflagbit ^ 1) == 1 {
		cflag = true
	}
	CpuSetFlags(ctx, nil, &nflag, &hflag, &cflag)

}

func procHalt(ctx *CpuContext) {
	ctx.Halted = true
}

func is16bit(rt regTypes) bool {
	return rt >= RT_AF
}
func CheckCondition(ctx *CpuContext) bool {
	z := CpuFlagZ()
	c := CpuFlagC()

	switch ctx.currentInst.Condition {
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

// Function pointer MAP
type InProc func(ctx *CpuContext)

var processors = make(map[InType]InProc)

func InitProcessors() {
	processors[IN_NONE] = procNone
	processors[IN_NOP] = procNop
	processors[IN_LD] = procLd
	processors[IN_JP] = procJp
	processors[IN_JR] = procJr
	processors[IN_CALL] = procCall
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
	processors[IN_AND] = procAnd
	processors[IN_XOR] = procXor
	processors[IN_OR] = procOr
	processors[IN_CP] = procCp
	processors[IN_CB] = procCb
	processors[IN_RLCA] = procRlca
	processors[IN_RRCA] = procRrca
	processors[IN_RLA] = procRla
	processors[IN_RRA] = procRra
	processors[IN_STOP] = procStop
	processors[IN_HALT] = procHalt
	processors[IN_DAA] = procDaa
	processors[IN_CPL] = procCpl
	processors[IN_SCF] = procScf
	processors[IN_CCF] = procCcf
	processors[IN_EI] = procEi
}

// fix this to return properly
func InstGetProccessor(intype InType) InProc {

	return processors[intype]

}

func CpuSetFlags(ctx *CpuContext, z *bool, n *bool, h *bool, c *bool) {
	if z != nil {
		ctx.Regs.F = BitSet(ctx.Regs.F, 7, z)
	}

	if n != nil {
		ctx.Regs.F = BitSet(ctx.Regs.F, 6, n)
	}

	if h != nil {
		ctx.Regs.F = BitSet(ctx.Regs.F, 5, h)
	}

	if c != nil {
		ctx.Regs.F = BitSet(ctx.Regs.F, 4, c)
	}
}
