package cpu

import (
	"app/internal/common"
	"app/internal/logger"
	"app/internal/memory"
)

var (
	debugLdHlCount  int
	debugLdSpCount  int
	debugAddSpCount int
	debugIncSpCount int
	debugDaaCount   int
)

func (c *CpuContext) ReadRegHL() uint16 {
	return (uint16(c.Regs.H) << 8) | uint16(c.Regs.L)
}

func (c *CpuContext) WriteRegHL(value uint16) {
	c.Regs.H = uint8(value >> 8)
	c.Regs.L = uint8(value & 0xFF)
}

func (c *CpuContext) ReadRegBC() uint16 {
	return (uint16(c.Regs.B) << 8) | uint16(c.Regs.C)
}

func (c *CpuContext) WriteRegBC(value uint16) {
	c.Regs.B = uint8(value >> 8)
	c.Regs.C = uint8(value & 0xFF)
}

func (c *CpuContext) ReadRegDE() uint16 {
	return (uint16(c.Regs.D) << 8) | uint16(c.Regs.E)
}

func (c *CpuContext) WriteRegDE(value uint16) {
	c.Regs.D = uint8(value >> 8)
	c.Regs.E = uint8(value & 0xFF)
}

func ProcINC16(cpu *CpuContext, reg16 *uint16) {
	*reg16++
	Cm.IncreaseCycle(1) // 16-bit increment takes 2 cycles total
}

func ProcDEC16(cpu *CpuContext, reg16 *uint16) {
	*reg16--
	Cm.IncreaseCycle(1) // 16-bit decrement takes 2 cycles total
}

func ProcADD_HL(cpu *CpuContext, value uint16) {
	hl := cpu.ReadRegHL()
	result := uint32(hl) + uint32(value)

	// Set flags for 16-bit ADD (matches reference implementation)
	z := false // Zero flag not affected by ADD HL
	n := false // Set to 0 for addition
	h := (hl&0x0FFF)+(value&0x0FFF) > 0x0FFF
	c := result > 0xFFFF

	CpuSetFlags(cpu, &z, &n, &h, &c)
	cpu.WriteRegHL(uint16(result & 0xFFFF))
	Cm.IncreaseCycle(1) // ADD HL takes 2 cycles total
}

func procNone(ctx *CpuContext) {
	logger.Fatal("Invalid Instruction!")
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
	if int(reg) >= len(rtLookup) {
		return RT_NONE
	}
	return rtLookup[reg]
}

func procNop(ctx *CpuContext) {
	// NOP: No operation. Correct.
}

func procLd(ctx *CpuContext) {
	if ctx.DestIsMem {
		if is16bit(ctx.currentInst.Reg2) {
			logger.Debug("LD mem16: opcode=%02X dest=%04X fetched=%04X srcReg=%d spNow=%04X", ctx.CurOpCode, ctx.MemDest, ctx.FetchedData, ctx.currentInst.Reg2, CpuRegRead(RT_SP))
			// Fault: 16-bit memory writes are rare (only LD (a16),SP). Make sure this is only used for correct instructions.
			Cm.IncreaseCycle(1)
			memory.BusCtx().BusWrite16(ctx.MemDest, ctx.FetchedData)
		} else {
			memory.BusCtx().BusWrite(ctx.MemDest, byte(ctx.FetchedData))
		}
		Cm.IncreaseCycle(1)
		return
	}

	if ctx.currentInst.Mode == AM_HL_SPR {
		// LD HL,SP+e8: apply signed 8-bit offset fetched during decode.
		offset := int8(ctx.FetchedData)
		sp := CpuRegRead(RT_SP)
		result := uint16(int32(sp) + int32(offset))
		logger.Debug("LD HL,SP+e8 executed: SP=%04X offset=%d result=%04X", sp, offset, result)

		offsetSigned := uint16(int16(offset))
		xorTerm := sp ^ offsetSigned ^ result
		h := (xorTerm & 0x0010) != 0
		c := (xorTerm & 0x0100) != 0
		expectedH := ((sp & 0x000F) + (uint16(byte(offset)) & 0x000F)) > 0x000F
		expectedC := ((sp & 0x00FF) + uint16(byte(offset))) > 0x00FF
		if h != expectedH || c != expectedC {
			logger.Warn("LD HL,SP+e8 flag mismatch: SP=%04X offset=%d result=%04X H=%t/%t C=%t/%t", sp, offset, result, h, expectedH, c, expectedC)
		}

		z := false
		n := false

		CpuSetFlags(ctx, &z, &n, &h, &c)
		flags := CpuRegRead(RT_F)
		if flags&0xC0 != 0 {
			logger.Warn("LD HL,SP+e8 unexpected Z/N flags: F=%02X SP=%04X offset=%d result=%04X", flags, sp, offset, result)
		}
		if offset == -1 && debugLdHlCount < 16 {
			debugLdHlCount++
			logger.Debug("LD HL,SP+e8 debug: SP=%04X result=%04X H=%t C=%t F=%02X", sp, result, h, c, flags)
		}
		CpuSetReg(RT_HL, result)

		if spAfter := CpuRegRead(RT_SP); spAfter != sp {
			logger.Warn("LD HL,SP+e8 mutated SP unexpectedly: before=%04X after=%04X offset=%d", sp, spAfter, offset)
		}

		Cm.IncreaseCycle(1)
		return
	}

	// Fault: For 16-bit LD r,nn, FetchedData should be 16 bits. For 8-bit LD, should mask to 8 bits.
	if ctx.currentInst.Mode == AM_R_MR && ctx.currentInst.Reg1 == RT_A && ctx.currentInst.Reg2 == RT_BC {
		logger.Debug("LD A,(BC) debug: BC=%04X value=%02X", CpuRegRead(RT_BC), byte(ctx.FetchedData))
	}
	CpuSetReg(ctx.currentInst.Reg1, ctx.FetchedData)

	if ctx.currentInst.Mode == AM_R_R && ctx.currentInst.Reg1 == RT_SP && ctx.currentInst.Reg2 == RT_HL {
		if debugLdSpCount < 32 {
			debugLdSpCount++
			logger.Debug("LD SP,HL debug: HL=%04X -> SP=%04X", CpuRegRead(RT_HL), CpuRegRead(RT_SP))
		}
	}
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
			z := result == 0
			n := false
			h := false
			c := (regval & 0x80) != 0

			CpuSetReg8(reg, result)
			CpuSetFlags(ctx, &z, &n, &h, &c)
			return

		case 1:
			// RRC r
			result := (regval >> 1) | (regval << 7)
			z := result == 0
			n := false
			h := false
			c := (regval & 0x01) != 0

			CpuSetReg8(reg, result)
			CpuSetFlags(ctx, &z, &n, &h, &c)
			return

		case 2:
			// RL r
			old := regval
			regval = regval << 1
			if CpuFlagC() {
				regval |= 1
			}
			z := regval == 0
			n := false
			h := false
			c := (old & 0x80) != 0

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &z, &n, &h, &c)
			return

		case 3:
			// RR r
			old := regval
			regval = regval >> 1
			if CpuFlagC() {
				regval |= 0x80
			}
			z := regval == 0
			n := false
			h := false
			c := (old & 0x01) != 0

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &z, &n, &h, &c)
			return

		case 4:
			// SLA r
			old := regval
			regval <<= 1
			z := regval == 0
			n := false
			h := false
			c := (old & 0x80) != 0

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &z, &n, &h, &c)
			return

		case 5:
			// SRA r
			old := regval
			msb := regval & 0x80
			regval = (regval >> 1) | msb
			z := regval == 0
			n := false
			h := false
			c := (old & 0x01) != 0

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &z, &n, &h, &c)
			return

		case 6:
			// SWAP r
			regval = ((regval & 0xF0) >> 4) | ((regval & 0x0F) << 4)
			z := regval == 0
			n := false
			h := false
			c := false

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &z, &n, &h, &c)
			return

		case 7:
			// SRL r
			old := regval
			regval >>= 1
			z := regval == 0
			n := false
			h := false
			c := (old & 0x01) != 0

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &z, &n, &h, &c)
			return
		}

	case 1:
		// BIT b, r
		z := (regval & (1 << bit)) == 0
		n := false
		h := true
		CpuSetFlags(ctx, &z, &n, &h, nil)
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
		logger.Fatal("ERROR: INVALID CB: %02X", op)
	}
}

// Could be a problem with casting here
func procAnd(ctx *CpuContext) {
	ctx.Regs.A &= byte(ctx.FetchedData)
	var z = ctx.Regs.A == 0
	var n = false
	var h = true // H always set for AND
	var c = false

	CpuSetFlags(ctx, &z, &n, &h, &c)
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

	z := false
	n := false
	h := false
	c := b == 1

	CpuSetFlags(ctx, &z, &n, &h, &c)
}

func procRla(ctx *CpuContext) {
	u := ctx.Regs.A
	c := (u & 0x80) != 0

	ctx.Regs.A <<= 1
	if CpuFlagC() {
		ctx.Regs.A |= 0x01
	}

	z := false
	n := false
	h := false

	CpuSetFlags(ctx, &z, &n, &h, &c)
}

func procRra(ctx *CpuContext) {
	bit0 := ctx.Regs.A & 0x01
	ctx.Regs.A >>= 1
	if CpuFlagC() {
		ctx.Regs.A |= 0x80
	}

	z := false
	n := false
	h := false
	c := bit0 == 1

	CpuSetFlags(ctx, &z, &n, &h, &c)
}

func procXor(ctx *CpuContext) {
	ctx.Regs.A ^= byte(ctx.FetchedData & 0xFF)

	z := ctx.Regs.A == 0
	n := false
	h := false
	c := false

	CpuSetFlags(ctx, &z, &n, &h, &c)
}

func procOr(ctx *CpuContext) {
	ctx.Regs.A |= byte(ctx.FetchedData & 0xFF)

	z := ctx.Regs.A == 0
	n := false
	h := false
	c := false

	CpuSetFlags(ctx, &z, &n, &h, &c)
}

func procCp(ctx *CpuContext) {
	a := ctx.Regs.A
	operand := byte(ctx.FetchedData & 0xFF)
	result := int16(a) - int16(operand)

	z := (result == 0)
	n := true
	h := (int(a&0x0F) - int(operand&0x0F)) < 0
	c := result < 0

	CpuSetFlags(ctx, &z, &n, &h, &c)
}

func procDi(ctx *CpuContext) {
	// DI: Disable interrupts
	ctx.IntMasterEnabled = false
}

func procEi(ctx *CpuContext) {
	// EI: Enable interrupts after next instruction
	logger.Debug("procEi invoked at PC=%04X", ctx.Regs.Pc)
	ctx.enablingIme = true
}

func procPop(ctx *CpuContext) {
	// POP rr: Pop two bytes from stack into register pair
	n := StackPop16()
	Cm.IncreaseCycle(2) // StackPop16 already handles the memory accesses
	CpuSetReg(ctx.currentInst.Reg1, n)
	if ctx.currentInst.Reg1 == RT_AF {
		// Lower 4 bits of F always zero
		CpuSetReg(ctx.currentInst.Reg1, n&0xFFF0)
	}
}

func procPush(ctx *CpuContext) {
	// PUSH rr: Push register pair onto stack
	value := CpuRegRead(ctx.currentInst.Reg1)
	Cm.IncreaseCycle(1)
	StackPush16(value)
	Cm.IncreaseCycle(1)
}

func goToAddr(ctx *CpuContext, addr uint16, pushpc bool) {
	if CheckCondition(ctx) {
		if pushpc {
			// CALL or RST: push PC before jump
			// Stack is LIFO: push HIGH byte first, then LOW byte
			value := ctx.Regs.Pc
			hi := byte((value >> 8) & 0xFF)
			lo := byte(value & 0xFF)
			Cm.IncreaseCycle(1)
			StackPush(hi)
			Cm.IncreaseCycle(1)
			StackPush(lo)
		}
		ctx.Regs.Pc = addr
		Cm.IncreaseCycle(1)
	}
}

func procJp(ctx *CpuContext) {
	// JP nn or JP cc,nn: Jump to address
	if CheckCondition(ctx) {
		ctx.Regs.Pc = ctx.FetchedData
		Cm.IncreaseCycle(1) // Jump cycle
	}
}

// Jump relative
func procJr(ctx *CpuContext) {
	if CheckCondition(ctx) {
		rel := int8(ctx.FetchedData)
		addr := uint16(int32(ctx.Regs.Pc) + int32(rel))
		ctx.Regs.Pc = addr
		Cm.IncreaseCycle(1) // Jump cycle
	}
}

func procCall(ctx *CpuContext) {
	// CALL nn or CALL cc,nn: Call subroutine
	if CheckCondition(ctx) {
		// Push current PC to stack
		Cm.IncreaseCycle(1)
		StackPush16(ctx.Regs.Pc)
		// Jump to new address
		ctx.Regs.Pc = ctx.FetchedData
		Cm.IncreaseCycle(1)
	}
}

// 0000002D problem here with fetched data, why stack push wrong
func procRet(ctx *CpuContext) {
	// RET or RET cc: Return from subroutine
	if ctx.currentInst.Condition != CT_NONE {
		Cm.IncreaseCycle(1) // Conditional check takes 1 cycle
		if !CheckCondition(ctx) {
			return // Condition not met, don't return
		}
	}

	// Pop return address from stack
	ctx.Regs.Pc = StackPop16()
	Cm.IncreaseCycle(2) // StackPop16 handles memory access cycles
	Cm.IncreaseCycle(1) // Jump cycle
}

func procRst(ctx *CpuContext) {
	// RST vec: Call fixed address (push PC, jump to vec)
	// Push current PC to stack
	Cm.IncreaseCycle(1)
	StackPush16(ctx.Regs.Pc)
	// Jump to RST vector
	ctx.Regs.Pc = uint16(ctx.currentInst.Param)
	Cm.IncreaseCycle(1)
}

func procReti(ctx *CpuContext) {
	// RETI: Return and enable interrupts (enable FIRST like reference implementation)
	ctx.IntMasterEnabled = true
	procRet(ctx)
}

func procLdh(ctx *CpuContext) {
	// LDH (a8),A or LDH A,(a8): High RAM I/O
	if ctx.currentInst.Reg1 == RT_A {
		// LDH A,(a8) - read from high RAM
		addr := 0xFF00 | (ctx.FetchedData & 0xFF)
		ctx.Regs.A = memory.BusCtx().BusRead(addr)
	} else {
		// LDH (a8),A - write to high RAM
		// For AM_A8_R, mem_dest is already set in fetch_data
		memory.BusCtx().BusWrite(ctx.MemDest, ctx.Regs.A)
	}
	Cm.IncreaseCycle(1)
}

func procInc(ctx *CpuContext) {
	if ctx.currentInst.Mode == AM_MR {
		addr := CpuRegRead(RT_HL)
		old := memory.BusCtx().BusRead(addr)
		value := uint16(old) + 1
		memory.BusCtx().BusWrite(addr, byte(value&0xFF))
		Cm.IncreaseCycle(1)

		z := (value & 0xFF) == 0
		n := false
		h := ((old & 0x0F) + 1) > 0x0F

		CpuSetFlags(ctx, &z, &n, &h, nil)
		return
	}

	if is16bit(ctx.currentInst.Reg1) {
		before := CpuRegRead(ctx.currentInst.Reg1)
		value := before + 1
		CpuSetReg(ctx.currentInst.Reg1, value)
		if ctx.currentInst.Reg1 == RT_SP && debugIncSpCount < 32 {
			debugIncSpCount++
			regs := CpuGetRegs()
			logger.Debug("INC SP debug: before=%04X after=%04X AF=%02X%02X BC=%02X%02X DE=%02X%02X HL=%02X%02X", before, regs.Sp, regs.A, regs.F, regs.B, regs.C, regs.D, regs.E, regs.H, regs.L)
		}
		Cm.IncreaseCycle(1)
		return
	}

	old := byte(CpuRegRead(ctx.currentInst.Reg1) & 0xFF)
	value := uint16(old) + 1
	CpuSetReg(ctx.currentInst.Reg1, value)

	z := (value & 0xFF) == 0
	n := false
	h := ((old & 0x0F) + 1) > 0x0F

	CpuSetFlags(ctx, &z, &n, &h, nil)
}

func procDec(ctx *CpuContext) {
	if ctx.currentInst.Mode == AM_MR {
		addr := CpuRegRead(RT_HL)
		old := memory.BusCtx().BusRead(addr)
		value := uint16(old) - 1
		memory.BusCtx().BusWrite(addr, byte(value&0xFF))
		Cm.IncreaseCycle(1)

		z := (value & 0xFF) == 0
		n := true
		h := (old & 0x0F) == 0

		CpuSetFlags(ctx, &z, &n, &h, nil)
		return
	}

	if is16bit(ctx.currentInst.Reg1) {
		value := CpuRegRead(ctx.currentInst.Reg1) - 1
		CpuSetReg(ctx.currentInst.Reg1, value)
		Cm.IncreaseCycle(1)
		return
	}

	old := byte(CpuRegRead(ctx.currentInst.Reg1) & 0xFF)
	value := uint16(old) - 1
	CpuSetReg(ctx.currentInst.Reg1, value)

	z := (value & 0xFF) == 0
	n := true
	h := (old & 0x0F) == 0

	CpuSetFlags(ctx, &z, &n, &h, nil)
}

func procSub(ctx *CpuContext) {
	a := ctx.Regs.A
	operand := byte(ctx.FetchedData & 0xFF)
	result := int16(a) - int16(operand)
	ctx.Regs.A = byte(result & 0xFF)

	z := ctx.Regs.A == 0
	n := true
	h := (a & 0x0F) < (operand & 0x0F) // H set if borrow from bit 4
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
	result := int16(a) - int16(operand) - int16(carry)
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
	if ctx.currentInst.Reg1 == RT_SP && ctx.currentInst.Mode == AM_R_D8 {
		// ADD SP, e8
		offset := int8(ctx.FetchedData)
		sp := CpuRegRead(RT_SP)
		result := uint16(int32(sp) + int32(offset))
		offsetSigned := uint16(int16(offset))
		xorTerm := sp ^ offsetSigned ^ result
		h := (xorTerm & 0x0010) != 0
		c := (xorTerm & 0x0100) != 0

		z := false
		n := false

		CpuSetReg(RT_SP, result)
		CpuSetFlags(ctx, &z, &n, &h, &c)
		if debugAddSpCount < 32 {
			debugAddSpCount++
			flags := CpuRegRead(RT_F)
			expectedH := ((sp & 0x000F) + (uint16(byte(offset)) & 0x000F)) > 0x000F
			expectedC := ((sp & 0x00FF) + uint16(byte(offset))) > 0x00FF
			logger.Debug("ADD SP,e8 debug: SP=%04X offset=%d result=%04X H=%t/%t C=%t/%t F=%02X", sp, offset, result, h, expectedH, c, expectedC, flags)
		}
		Cm.IncreaseCycle(2)
		return
	}

	if is16bit(ctx.currentInst.Reg1) {
		// ADD HL,rr. Z not affected, N=0, H and C as per spec.
		val1 := CpuRegRead(ctx.currentInst.Reg1)
		val2 := CpuRegRead(ctx.currentInst.Reg2)
		result := val1 + val2

		h := ((val1 & 0x0FFF) + (val2 & 0x0FFF)) > 0x0FFF
		c := (uint32(val1) + uint32(val2)) > 0xFFFF

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
	// STOP: Enter low-power mode (not fully emulated here)
	logger.Debug("STOP instruction encountered; halting CPU")
	ctx.Halted = true
	//ctx.Stopped = true
}

func procDaa(ctx *CpuContext) {
	// DAA: Decimal adjust accumulator after addition/subtraction
	origA := ctx.Regs.A
	origF := ctx.Regs.F

	a := origA
	n := CpuFlagN()
	h := CpuFlagH()
	c := CpuFlagC()

	carry := c
	var adjust byte

	if !n {
		if c || a > 0x99 {
			adjust |= 0x60
			carry = true
		}
		if h || (a&0x0F) > 0x09 {
			adjust |= 0x06
		}
		a = byte(uint16(a) + uint16(adjust))
	} else {
		if c {
			adjust |= 0x60
		}
		if h {
			adjust |= 0x06
		}
		a = byte(uint16(a) - uint16(adjust))
	}

	a &= 0xFF
	z := a == 0
	hClear := false

	ctx.Regs.A = a
	CpuSetFlags(ctx, &z, nil, &hClear, &carry)

	expectedF := byte(0)
	if z {
		expectedF |= 0x80
	}
	if (origF & 0x40) != 0 {
		expectedF |= 0x40
	}
	if carry {
		expectedF |= 0x10
	}

	if debugDaaCount < 64 {
		debugDaaCount++
		logger.Debug("DAA debug: PC=%04X N=%t H=%t C_in=%t adjust=%02X A_in=%02X -> A_out=%02X C_out=%t F_out=%02X", ctx.Regs.Pc, n, h, c, adjust, origA, ctx.Regs.A, carry, ctx.Regs.F)
	}

	if (ctx.Regs.F & 0xF0) != expectedF {
		logger.Warn("DAA flag mismatch: A_in=%02X F_in=%02X -> A_out=%02X F_out=%02X expectedF=%02X", origA, origF, ctx.Regs.A, ctx.Regs.F, expectedF)
	}
}

func procCpl(ctx *CpuContext) {
	// CPL: Complement accumulator
	ctx.Regs.A = ^ctx.Regs.A
	n := true
	h := true
	CpuSetFlags(ctx, nil, &n, &h, nil)
}

func procScf(ctx *CpuContext) {
	// SCF: Set carry flag
	n := false
	h := false
	c := true
	CpuSetFlags(ctx, nil, &n, &h, &c)
}

func procCcf(ctx *CpuContext) {
	// CCF: Complement carry flag
	c := !CpuFlagC()
	n := false
	h := false
	CpuSetFlags(ctx, nil, &n, &h, &c)
}

func procHalt(ctx *CpuContext) {
	// HALT: Enter low-power mode (not fully emulated here)
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
		ctx.Regs.F = common.BitSet(ctx.Regs.F, 7, *z)
	}

	if n != nil {
		ctx.Regs.F = common.BitSet(ctx.Regs.F, 6, *n)
	}

	if h != nil {
		ctx.Regs.F = common.BitSet(ctx.Regs.F, 5, *h)
	}

	if c != nil {
		ctx.Regs.F = common.BitSet(ctx.Regs.F, 4, *c)
	}
}
