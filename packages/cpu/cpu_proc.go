package cpu

import (
	"log"
	"pajalic.go.emulator/packages/memory"
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
			Cm.IncreaseCycle(1)
			memory.BusWrite16(ctx.MemDest, ctx.FetchedData)
		} else {
			memory.BusWrite(ctx.MemDest, byte(ctx.FetchedData))
		}
		Cm.IncreaseCycle(1)
		return
	}

	if ctx.currentInst.Mode == AM_HL_SPR {
		//        u8 hflag = (cpu_read_reg(ctx->cur_inst->reg_2) & 0xF) +
		//            (ctx->fetched_data & 0xF) >= 0x10;
		//
		//        u8 cflag = (cpu_read_reg(ctx->cur_inst->reg_2) & 0xFF) +
		//            (ctx->fetched_data & 0xFF) >= 0x100;
		//
		//        cpu_set_flags(ctx, 0, 0, hflag, cflag);
		//        cpu_set_reg(ctx->cur_inst->reg_1,
		//            cpu_read_reg(ctx->cur_inst->reg_2) + (int8_t)ctx->fetched_data);
		//
		//        return;

		//

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
	Cm.IncreaseCycle(1)

	if reg == RT_HL {
		Cm.IncreaseCycle(2)
	}

	switch bit_op {
	case 1:
		zflag := (regval & (1 << bit)) == 0
		nflag := false
		hflag := true
		//BIT , is this correct, should it be equals to zero
		CpuSetFlags(ctx, &zflag, &nflag, &hflag, nil)
		return

	case 2:
		//RST
		//Does ^ equal ~ in c
		regval &= ^(1 << bit)
		CpuSetReg8(reg, regval)
		return

	case 3:
		//SET
		regval |= 1 << bit
		CpuSetReg8(reg, regval)
		return
	}
	flagC := CpuFlagC()

	switch bit {
	case 0:
		{
			//RLC
			var setC = false
			var result = (regval << 1) & 0xFF

			// Set carry flag based on carry-out from MSB of regval
			if (regval & (1 << 7)) != 0 {
				setC = true
				result |= 1 // Move the MSB to the LSB
			}

			// Set zero flag
			zflag := result == 0

			// Negative and half carry flags are not set in this operation
			nflag := false
			hflag := false

			CpuSetReg8(reg, result)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &setC)
			return
		}

	case 1:
		{
			//RRC
			var old = regval
			regval >>= 1
			regval |= old << 7

			zflag := regval == 0
			nflag := false
			hflag := false
			cflag := old&1 == 1

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return
		}

	case 2:
		{
			//RL
			var old = regval
			regval <<= 1
			// byte to bool
			var bitCflag byte = 0

			if flagC {
				bitCflag = 1
			}
			regval |= bitCflag

			zflag := regval == 0
			nflag := false
			hflag := false
			cflag := (old & 0x80) != 0

			CpuSetReg8(reg, regval)
			// Clamp old between 0 and 1
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return
		}

	case 3:
		{
			//RR
			//            u8 old = reg_val;
			//            reg_val >>= 1;
			//
			//            reg_val |= (flagC << 7);
			//
			//            cpu_set_reg8(reg, reg_val);
			//            cpu_set_flags(ctx, !reg_val, false, false, old & 1);

			var old = regval
			regval >>= 1

			var bitCflag byte = 0
			if flagC {
				bitCflag = 1
			}

			// Set carry flag based on the least significant bit of the old value
			cflag := old&1 == 1

			// Set the most significant bit of regval based on the carry flag
			regval |= bitCflag << 7

			// Set zero flag
			zflag := regval == 0

			// Negative and half carry flags are not set in this operation
			nflag := false
			hflag := false

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return
		}

	case 4:
		{
			//SLA
			var old = regval
			regval <<= 1

			zflag := regval == 0
			nflag := false
			hflag := false
			cflag := (old & 0x80) != 0

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return
		}

	case 5:
		{
			//SRA
			// what is int8_t MSB should not change, is this true
			var u = byte(int8(regval) >> 1)

			zflag := u == 0
			nflag := false
			hflag := false
			cflag := regval&1 == 1
			CpuSetReg8(reg, u)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return
		}

	case 6:
		{
			//SWAP
			regval = ((regval & 0xF0) >> 4) | ((regval & 0xF) << 4)
			zflag := regval == 0
			nflag := false
			hflag := false
			cflag := false

			CpuSetReg8(reg, regval)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return
		}

	case 7:
		{
			//SRL
			var u = regval >> 1
			zflag := u == 0
			nflag := false
			hflag := false
			cflag := regval&1 == 1

			CpuSetReg8(reg, u)
			CpuSetFlags(ctx, &zflag, &nflag, &hflag, &cflag)
			return
		}
	}
	log.Fatal("ERROR: INVALID CB: %02X", op)

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

	//    u8 u = ctx->regs.a;
	//    bool c = (u >> 7) & 1;
	//    u = (u << 1) | c;
	//    ctx->regs.a = u;
	//
	//    cpu_set_flags(ctx, 0, 0, 0, c);
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
	Cm.IncreaseCycle(1)
	var hi = uint16(StackPop())
	Cm.IncreaseCycle(1)
	var n = (hi << 8) | lo
	CpuSetReg(ctx.currentInst.Reg1, n)

	if ctx.currentInst.Reg1 == RT_AF {
		CpuSetReg(ctx.currentInst.Reg1, n&0xFFF0)
	}
}

func procPush(ctx *CpuContext) {
	var hi = (CpuRegRead(ctx.currentInst.Reg1) >> 8) & 0xFF
	Cm.IncreaseCycle(1)
	StackPush(byte(hi))
	var lo = (CpuRegRead(ctx.currentInst.Reg1)) & 0xFF
	Cm.IncreaseCycle(1)
	StackPush(byte(lo))
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
	var rel = int8(ctx.FetchedData & 0xFF) //casting cause it might be negative
	var addr = ctx.Regs.Pc + uint16(rel)
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
		var lo = uint16(StackPop())
		Cm.IncreaseCycle(1)

		var hi = uint16(StackPop())
		Cm.IncreaseCycle(1)

		var n = (hi << 8) | lo
		ctx.Regs.Pc = n

		Cm.IncreaseCycle(1)
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
		CpuSetReg(ctx.currentInst.Reg1, uint16(memory.BusRead(0xFF00|ctx.FetchedData)))
	} else {
		memory.BusWrite(ctx.MemDest, ctx.Regs.A)
	}
	Cm.IncreaseCycle(1)
}

func procInc(ctx *CpuContext) {
	var val = CpuRegRead(ctx.currentInst.Reg1) + 1

	if is16bit(ctx.currentInst.Reg1) {
		Cm.IncreaseCycle(1)
	}

	if ctx.currentInst.Reg1 == RT_HL && ctx.currentInst.Mode == AM_MR {
		val = uint16(memory.BusRead(CpuRegRead(RT_HL))) + 1
		val &= 0xFF
		memory.BusWrite(CpuRegRead(RT_HL), byte(val))
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
		Cm.IncreaseCycle(1)
	}

	if ctx.currentInst.Reg1 == RT_HL && ctx.currentInst.Mode == AM_MR {
		val = uint16(memory.BusRead(CpuRegRead(RT_HL)) - 1)
		memory.BusWrite(CpuRegRead(RT_HL), byte(val))
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

	//    u16 val = cpu_read_reg(ctx->cur_inst->reg_1) - ctx->fetched_data;
	//
	//    int z = val == 0;
	//    int h = ((int)cpu_read_reg(ctx->cur_inst->reg_1) & 0xF) - ((int)ctx->fetched_data & 0xF) < 0;
	//    int c = ((int)cpu_read_reg(ctx->cur_inst->reg_1)) - ((int)ctx->fetched_data) < 0;
	//
	//    cpu_set_reg(ctx->cur_inst->reg_1, val);
	//    cpu_set_flags(ctx, z, 1, h, c);
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

	//    u16 u = ctx->fetched_data;
	//    u16 a = ctx->regs.a;
	//    u16 c = CPU_FLAG_C;
	//
	//    ctx->regs.a = (a + u + c) & 0xFF;
	//
	//    cpu_set_flags(ctx, ctx->regs.a == 0, 0,
	//        (a & 0xF) + (u & 0xF) + c > 0xF,
	//        a + u + c > 0xFF);
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
		Cm.IncreaseCycle(1)
	}

	if ctx.currentInst.Reg1 == RT_SP {
		val = uint32(CpuRegRead(ctx.currentInst.Reg1) + uint16(byte(ctx.FetchedData)))
	}

	z := (val & 0xFF) == 0

	h := (CpuRegRead(ctx.currentInst.Reg1)&0xF)+(ctx.FetchedData&0xF) >= 0x10
	c := (int)(CpuRegRead(ctx.currentInst.Reg1)&0xFF)+(int)(ctx.FetchedData&0xFF) >= 0x100
	n := false

	//    if (is_16bit) {
	//        z = -1;
	//        h = (cpu_read_reg(ctx->cur_inst->reg_1) & 0xFFF) + (ctx->fetched_data & 0xFFF) >= 0x1000;
	//        u32 n = ((u32)cpu_read_reg(ctx->cur_inst->reg_1)) + ((u32)ctx->fetched_data);
	//        c = n >= 0x10000;
	//    }

	if is16bit {
		h = (CpuRegRead(ctx.currentInst.Reg1)&0xFFF)+(ctx.FetchedData&0xFFF) >= 0x1000
		c = (uint32(CpuRegRead(ctx.currentInst.Reg1)))+uint32(ctx.FetchedData) >= 0x10000

	}

	//    if (ctx->cur_inst->reg_1 == RT_SP) {
	//        z = 0;
	//        h = (cpu_read_reg(ctx->cur_inst->reg_1) & 0xF) + (ctx->fetched_data & 0xF) >= 0x10;
	//        c = (int)(cpu_read_reg(ctx->cur_inst->reg_1) & 0xFF) + (int)(ctx->fetched_data & 0xFF) >= 0x100;
	//    }
	//
	//    cpu_set_reg(ctx->cur_inst->reg_1, val & 0xFFFF);
	//    cpu_set_flags(ctx, z, 0, h, c);

	//  u32 val = cpu_read_reg(ctx->cur_inst->reg_1) + ctx->fetched_data;
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
