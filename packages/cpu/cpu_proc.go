package cpu

import (
	"log"
	"pajalic.go.emulator/packages/pubsub"
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
	if reg >= 0b111 {
		return RT_NONE
	}

	return rtLookup[reg]
}

func procNop(cpucontext *CpuContext) {

}

func procLd(ctx *CpuContext) {
	if ctx.DestIsMem {
		if is16bit(ctx.currentInst.Reg2) {
			Cm.IncreaseCycle(1)
			pubsub.BusCtx().BusWrite16(ctx.MemDest, ctx.FetchedData)
		} else {
			pubsub.BusCtx().BusWrite(ctx.MemDest, byte(ctx.FetchedData))
		}
		Cm.IncreaseCycle(1)
		return
	}

	if ctx.currentInst.Mode == AM_HL_SPR {
		value := int8(ctx.FetchedData)
		sp := CpuRegRead(RT_SP)
		result := uint16(int32(sp) + int32(value))

		hflag := ((sp & 0x0F) + (uint16(value) & 0x0F)) > 0x0F
		cflag := ((sp & 0xFF) + (uint16(value) & 0xFF)) > 0xFF

		zflag := false
		nflag := false

		CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
		CpuSetReg(RT_HL, result)

		Cm.IncreaseCycle(1)
		return
	}

	CpuSetReg(ctx.currentInst.Reg1, ctx.FetchedData)
}

func procCb(ctx *CpuContext) {
	op := byte(ctx.FetchedData)
	reg := decodeReg(op & 0b111)
	bit := (op >> 3) & 0b111
	bitOp := (op >> 6) & 0b11
	regval := CpuRegRead8(reg)
	Cm.IncreaseCycle(1)

	if reg == RT_HL {
		Cm.IncreaseCycle(2)
	}

	switch bitOp {
	case 0:
		// Rotate and Shift Operations
		switch bit {
		case 0:
			// RLC r
			result := (regval << 1) | (regval >> 7)
			zflag := result == 0
			nflag := false
			hflag := false
			cflag := (regval & 0x80) != 0

			CpuSetReg8(reg, result)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return

		case 1:
			// RRC r
			result := (regval >> 1) | (regval << 7)
			zflag := result == 0
			nflag := false
			hflag := false
			cflag := (regval & 0x01) != 0

			CpuSetReg8(reg, result)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return

		case 2:
			// RL r
			old := regval
			regval = regval << 1
			if CpuFlagC() {
				regval |= 1
			}
			zflag := regval == 0
			nflag := false
			hflag := false
			cflag := (old & 0x80) != 0

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return

		case 3:
			// RR r
			old := regval
			regval = regval >> 1
			if CpuFlagC() {
				regval |= 0x80
			}
			zflag := regval == 0
			nflag := false
			hflag := false
			cflag := (old & 0x01) != 0

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return

		case 4:
			// SLA r
			old := regval
			regval <<= 1
			zflag := regval == 0
			nflag := false
			hflag := false
			cflag := (old & 0x80) != 0

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return

		case 5:
			// SRA r
			old := regval
			msb := regval & 0x80
			regval = (regval >> 1) | msb
			zflag := regval == 0
			nflag := false
			hflag := false
			cflag := (old & 0x01) != 0

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return

		case 6:
			// SWAP r
			regval = ((regval & 0xF0) >> 4) | ((regval & 0x0F) << 4)
			zflag := regval == 0
			nflag := false
			hflag := false
			cflag := false

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return

		case 7:
			// SRL r
			old := regval
			regval >>= 1
			zflag := regval == 0
			nflag := false
			hflag := false
			cflag := (old & 0x01) != 0

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return
		}

	case 1:
		// BIT b, r
		zflag := (regval & (1 << bit)) == 0
		nflag := false
		hflag := true
		CpuSetFlags(ctx, &zflag, &nflag, &hflag, nil)
		return

	case 2:
		// RES b, r
		regval &^= (1 << bit)
		CpuSetReg8(reg, regval)
		return

	case 3:
		// SET b, r
		regval |= 1 << bit
		CpuSetReg8(reg, regval)
		return

	default:
		log.Fatalf("ERROR: INVALID CB: %02X", op)
	}
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
	u := ctx.Regs.A
	c := (u & 0x80) != 0

	u = u << 1
	if c {
		u |= 0x01
	}

	ctx.Regs.A = u
	z := false
	n := false
	h := false

	CpuSetFlags(ctx, &z, &n, &h, &c)
}

func procRrca(ctx *CpuContext) {
	b := ctx.Regs.A & 0x01
	ctx.Regs.A >>= 1
	ctx.Regs.A |= b << 7

	zflag := false
	nflag := false
	hflag := false
	cflag := b == 1

	CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
}

func procRla(ctx *CpuContext) {
	u := ctx.Regs.A
	cflag := (u & 0x80) != 0

	ctx.Regs.A <<= 1
	if CpuFlagC() {
		ctx.Regs.A |= 0x01
	}

	zflag := false
	nflag := false
	hflag := false

	CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
}

func procRra(ctx *CpuContext) {
	bit0 := ctx.Regs.A & 0x01
	ctx.Regs.A >>= 1
	if CpuFlagC() {
		ctx.Regs.A |= 0x80
	}

	zflag := false
	nflag := false
	hflag := false
	cflag := bit0 == 1

	CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
}

func procXor(ctx *CpuContext) {
	ctx.Regs.A ^= byte(ctx.FetchedData & 0xFF)

	zflag := ctx.Regs.A == 0
	nflag := false
	hflag := false
	cflag := false

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
	a := ctx.Regs.A
	operand := byte(ctx.FetchedData & 0xFF)

	zflag := a == operand
	nflag := true
	hflag := (a & 0x0F) < (operand & 0x0F)
	cflag := a < operand

	CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
}

func procDi(ctx *CpuContext) {
	ctx.IntMasterEnabled = false
}

func procEi(ctx *CpuContext) {
	ctx.enablingIme = true
}

func procPop(ctx *CpuContext) {
	lo := uint16(StackPop())
	Cm.IncreaseCycle(1)
	hi := uint16(StackPop())
	Cm.IncreaseCycle(1)
	n := (hi << 8) | lo
	CpuSetReg(ctx.currentInst.Reg1, n)

	if ctx.currentInst.Reg1 == RT_AF {
		CpuSetReg(ctx.currentInst.Reg1, n&0xFFF0)
	}
}

func procPush(ctx *CpuContext) {
	hi := byte((CpuRegRead(ctx.currentInst.Reg1) >> 8) & 0xFF)
	Cm.IncreaseCycle(1)
	StackPush(hi)
	lo := byte(CpuRegRead(ctx.currentInst.Reg1) & 0xFF)
	Cm.IncreaseCycle(1)
	StackPush(lo)
	Cm.IncreaseCycle(1)
}

func goToAddr(ctx *CpuContext, addr uint16, pushpc bool) {
	if CheckCondition(ctx) {
		if pushpc {
			Cm.IncreaseCycle(2)
			StackPush16(ctx.Regs.Pc)
		}
		ctx.Regs.Pc = addr
		Cm.IncreaseCycle(1)
	}
}

func procJp(ctx *CpuContext) {
	goToAddr(ctx, ctx.FetchedData, false)
}

// Jump relative
func procJr(ctx *CpuContext) {
	rel := int8(ctx.FetchedData & 0xFF)
	addr := ctx.Regs.Pc + uint16(rel)
	goToAddr(ctx, addr, false)
}

func procCall(ctx *CpuContext) {
	goToAddr(ctx, ctx.FetchedData, true)
}

// 0000002D problem here with fetched data, why stack push wrong
func procRet(ctx *CpuContext) {
	if ctx.currentInst.Condition != CT_NONE {
		Cm.IncreaseCycle(1)
	}

	if CheckCondition(ctx) {
		lo := uint16(StackPop())
		Cm.IncreaseCycle(1)
		hi := uint16(StackPop())
		Cm.IncreaseCycle(1)
		n := (hi << 8) | lo
		ctx.Regs.Pc = n
		Cm.IncreaseCycle(1)
	}
}

func procRst(ctx *CpuContext) {
	goToAddr(ctx, uint16(ctx.currentInst.Param), true)
}

func procReti(ctx *CpuContext) {
	procRet(ctx)
	ctx.IntMasterEnabled = true
}

func procLdh(ctx *CpuContext) {
	if ctx.currentInst.Reg1 == RT_A {
		value := pubsub.BusCtx().BusRead(0xFF00 | ctx.FetchedData)
		ctx.Regs.A = value
	} else {
		pubsub.BusCtx().BusWrite(0xFF00|ctx.FetchedData, ctx.Regs.A)
	}
	Cm.IncreaseCycle(1)
}

func procInc(ctx *CpuContext) {
	if is16bit(ctx.currentInst.Reg1) {
		value := CpuRegRead(ctx.currentInst.Reg1) + 1
		CpuSetReg(ctx.currentInst.Reg1, value)
		Cm.IncreaseCycle(1)
		return
	}

	var value uint16
	if ctx.currentInst.Mode == AM_MR {
		addr := CpuRegRead(RT_HL)
		data := pubsub.BusCtx().BusRead(addr)
		value = uint16(data) + 1
		pubsub.BusCtx().BusWrite(addr, byte(value&0xFF))
		Cm.IncreaseCycle(1)
	} else {
		value = CpuRegRead(ctx.currentInst.Reg1) + 1
		CpuSetReg(ctx.currentInst.Reg1, value)
	}

	z := (value & 0xFF) == 0
	n := false
	h := ((value-1)&0x0F)+1 > 0x0F

	CpuSetFlags(ctx, &z, &n, &h, nil)
}

func procDec(ctx *CpuContext) {
	if is16bit(ctx.currentInst.Reg1) {
		value := CpuRegRead(ctx.currentInst.Reg1) - 1
		CpuSetReg(ctx.currentInst.Reg1, value)
		Cm.IncreaseCycle(1)
		return
	}

	var value uint16
	if ctx.currentInst.Mode == AM_MR {
		addr := CpuRegRead(RT_HL)
		data := pubsub.BusCtx().BusRead(addr)
		value = uint16(data) - 1
		pubsub.BusCtx().BusWrite(addr, byte(value&0xFF))
		Cm.IncreaseCycle(1)
	} else {
		value = CpuRegRead(ctx.currentInst.Reg1) - 1
		CpuSetReg(ctx.currentInst.Reg1, value)
	}

	z := (value & 0xFF) == 0
	n := true
	h := ((value + 1) & 0x0F) == 0x00

	CpuSetFlags(ctx, &z, &n, &h, nil)
}

func procSub(ctx *CpuContext) {
	a := ctx.Regs.A
	operand := byte(ctx.FetchedData & 0xFF)
	result := uint16(a) - uint16(operand)
	ctx.Regs.A = byte(result & 0xFF)

	z := ctx.Regs.A == 0
	n := true
	h := (int(a&0x0F) - int(operand&0x0F)) < 0
	c := a < operand

	CpuSetFlags(ctx, &z, &n, &h, &c)
}

func procSbc(ctx *CpuContext) {
	a := ctx.Regs.A
	operand := byte(ctx.FetchedData & 0xFF)
	carry := byte(0)
	if CpuFlagC() {
		carry = 1
	}
	result := uint16(a) - uint16(operand) - uint16(carry)
	ctx.Regs.A = byte(result & 0xFF)

	z := ctx.Regs.A == 0
	n := true
	h := (int(a&0x0F) - int(operand&0x0F) - int(carry)) < 0
	c := int(a)-int(operand)-int(carry) < 0

	CpuSetFlags(ctx, &z, &n, &h, &c)
}

func procAdc(ctx *CpuContext) {
	a := ctx.Regs.A
	operand := byte(ctx.FetchedData & 0xFF)
	carry := byte(0)
	if CpuFlagC() {
		carry = 1
	}
	result := uint16(a) + uint16(operand) + uint16(carry)
	ctx.Regs.A = byte(result & 0xFF)

	z := ctx.Regs.A == 0
	n := false
	h := ((a & 0x0F) + (operand & 0x0F) + carry) > 0x0F
	c := result > 0xFF

	CpuSetFlags(ctx, &z, &n, &h, &c)
}

// Bool to Int problematik som sker i C men ej i GOlang
func procAdd(ctx *CpuContext) {
	if ctx.currentInst.Reg1 == RT_SP && ctx.currentInst.Mode == AM_D8 {
		value := int8(ctx.FetchedData)
		sp := CpuRegRead(RT_SP)
		result := uint16(int32(sp) + int32(value))

		h := ((sp & 0x0F) + (uint16(value) & 0x0F)) > 0x0F
		c := ((sp & 0xFF) + (uint16(value) & 0xFF)) > 0xFF

		z := false
		n := false

		CpuSetReg(RT_SP, result)
		CpuSetFlags(ctx, &z, &n, &h, &c)
		Cm.IncreaseCycle(2)
		return
	}

	if is16bit(ctx.currentInst.Reg1) {
		val1 := CpuRegRead(ctx.currentInst.Reg1)
		val2 := CpuRegRead(ctx.currentInst.Reg2)
		result := val1 + val2

		h := ((val1 & 0x0FFF) + (val2 & 0x0FFF)) > 0x0FFF
		c := result > 0xFFFF

		n := false

		CpuSetReg(ctx.currentInst.Reg1, result&0xFFFF)
		CpuSetFlags(ctx, nil, &n, &h, &c)
		Cm.IncreaseCycle(1)
		return
	}

	a := ctx.Regs.A
	operand := byte(ctx.FetchedData & 0xFF)
	result := uint16(a) + uint16(operand)
	ctx.Regs.A = byte(result & 0xFF)

	z := ctx.Regs.A == 0
	n := false
	h := ((a & 0x0F) + (operand & 0x0F)) > 0x0F
	c := result > 0xFF

	CpuSetFlags(ctx, &z, &n, &h, &c)
}

func procStop(ctx *CpuContext) {
	log.Fatal("STOPPING!")
}

func procDaa(ctx *CpuContext) {
	var adjust byte = 0
	a := ctx.Regs.A
	c := CpuFlagC()

	if !CpuFlagN() {
		if CpuFlagH() || (a&0x0F) > 0x09 {
			adjust |= 0x06
		}
		if c || a > 0x99 {
			adjust |= 0x60
			c = true
		}
		a += adjust
	} else {
		if CpuFlagH() {
			adjust |= 0x06
		}
		if c {
			adjust |= 0x60
		}
		a -= adjust
	}

	a &= 0xFF
	z := a == 0
	h := false

	ctx.Regs.A = a
	CpuSetFlags(ctx, &z, nil, &h, &c)
}

func procCpl(ctx *CpuContext) {
	ctx.Regs.A = ^ctx.Regs.A
	nflag := true
	hflag := true

	CpuSetFlags(ctx, nil, &nflag, &hflag, nil)
}

func procScf(ctx *CpuContext) {
	nflag := false
	hflag := false
	cflag := true

	CpuSetFlags(ctx, nil, &nflag, &hflag, &cflag)
}
func procCcf(ctx *CpuContext) {
	cflag := !CpuFlagC()
	nflag := false
	hflag := false

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
		ctx.Regs.F = BitSet(ctx.Regs.F, 7, *z)
	}

	if n != nil {
		ctx.Regs.F = BitSet(ctx.Regs.F, 6, *n)
	}

	if h != nil {
		ctx.Regs.F = BitSet(ctx.Regs.F, 5, *h)
	}

	if c != nil {
		ctx.Regs.F = BitSet(ctx.Regs.F, 4, *c)
	}
}
