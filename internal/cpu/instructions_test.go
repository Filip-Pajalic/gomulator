package cpu

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"app/internal/memory"
)

func TestInstructionTableMatchesGBDevOpcodeSpec(t *testing.T) {
	InitInstructions()

	want := loadGBDevBaseInstructions(t)
	for opcode, expected := range want {
		t.Run(fmt.Sprintf("%02X", opcode), func(t *testing.T) {
			got := *instructionByOpcode(byte(opcode))
			if got != expected {
				t.Fatalf("opcode %02X = %+v, want %+v", opcode, got, expected)
			}
		})
	}
}

func TestAllBaseInstructionsHaveProcessors(t *testing.T) {
	InitInstructions()
	InitProcessors()

	for opcode := range inst {
		instruction := instructionByOpcode(byte(opcode))
		if instruction.Type == IN_NONE {
			continue
		}

		if processor := InstGetProccessor(instruction.Type); processor == nil {
			t.Fatalf("opcode %02X (%s) has no processor", opcode, getInstructionName(instruction.Type))
		}
	}
}

func TestAllBaseInstructionsCanStep(t *testing.T) {
	want := loadGBDevBaseInstructions(t)

	for opcode, expected := range want {
		if expected.Type == IN_NONE {
			continue
		}

		t.Run(fmt.Sprintf("%02X", opcode), func(t *testing.T) {
			ctx := newInstructionTestCPU()
			seedStepState(ctx, byte(opcode))

			if !ctx.Step() {
				t.Fatalf("step returned false")
			}

			if ctx.Regs.F&0x0F != 0 {
				t.Fatalf("lower flag nibble = %02X, want 00", ctx.Regs.F&0x0F)
			}
		})
	}
}

func TestCBPrefixedInstructionsExecuteAllOpcodes(t *testing.T) {
	opcodes := loadGBDevOpcodes(t)

	for opcode := 0; opcode <= 0xFF; opcode++ {
		t.Run(fmt.Sprintf("CB_%02X", opcode), func(t *testing.T) {
			assertGBDevCBSpec(t, opcodes.CBPrefixed[fmt.Sprintf("0x%02X", opcode)], byte(opcode))

			ctx := newInstructionTestCPU()
			ctx.currentInst = &Instruction{Type: IN_CB}
			ctx.FetchedData = uint16(opcode)
			ctx.Regs.F = 0x10

			reg := decodeReg(byte(opcode) & 0x07)
			initial := cbInitialValue(byte(opcode))
			writeCBTarget(reg, initial)

			wantValue, wantFlags := expectedCBResult(byte(opcode), initial, true)
			procCb(ctx)

			if got := readCBTarget(reg); got != wantValue {
				t.Fatalf("target value = %02X, want %02X", got, wantValue)
			}

			if got := ctx.Regs.F; got != wantFlags {
				t.Fatalf("flags = %02X, want %02X", got, wantFlags)
			}
		})
	}
}

func TestEIDelaysIMEUntilAfterFollowingInstruction(t *testing.T) {
	ctx := newInstructionTestCPU()
	ctx.Regs.Pc = 0x0100
	ctx.Regs.Sp = 0xD000
	ctx.IntFlags = byte(IT_TIMER)

	bus := memory.BusCtx()
	bus.BusWrite(0xFFFF, byte(IT_TIMER))
	bus.BusWrite(0x0100, 0xFB) // EI
	bus.BusWrite(0x0101, 0x00) // NOP

	if !ctx.Step() {
		t.Fatalf("EI step returned false")
	}
	if ctx.IntMasterEnabled {
		t.Fatalf("IME enabled immediately after EI")
	}
	if !ctx.enablingIme || ctx.imeEnableDelay != 1 {
		t.Fatalf("EI delay state = enabling:%t delay:%d, want enabling:true delay:1", ctx.enablingIme, ctx.imeEnableDelay)
	}
	if ctx.Regs.Pc != 0x0101 {
		t.Fatalf("PC after EI = %04X, want 0101", ctx.Regs.Pc)
	}

	if !ctx.Step() {
		t.Fatalf("NOP step returned false")
	}
	if ctx.IntMasterEnabled {
		t.Fatalf("IME still enabled after servicing interrupt")
	}
	if ctx.IntFlags&byte(IT_TIMER) != 0 {
		t.Fatalf("timer interrupt flag was not cleared: %02X", ctx.IntFlags)
	}
	if ctx.Regs.Pc != 0x0050 {
		t.Fatalf("PC after delayed interrupt = %04X, want 0050", ctx.Regs.Pc)
	}
	if ctx.Regs.Sp != 0xCFFE {
		t.Fatalf("SP after interrupt = %04X, want CFFE", ctx.Regs.Sp)
	}
	if lo, hi := bus.BusRead(0xCFFE), bus.BusRead(0xCFFF); lo != 0x02 || hi != 0x01 {
		t.Fatalf("pushed return address = %02X%02X, want 0102", hi, lo)
	}
}

func TestDICancelsPendingEIDelay(t *testing.T) {
	ctx := newInstructionTestCPU()
	ctx.Regs.Pc = 0x0100
	ctx.Regs.Sp = 0xD000
	ctx.IntFlags = byte(IT_TIMER)

	bus := memory.BusCtx()
	bus.BusWrite(0xFFFF, byte(IT_TIMER))
	bus.BusWrite(0x0100, 0xFB) // EI
	bus.BusWrite(0x0101, 0xF3) // DI
	bus.BusWrite(0x0102, 0x00) // NOP

	if !ctx.Step() {
		t.Fatalf("EI step returned false")
	}
	if !ctx.Step() {
		t.Fatalf("DI step returned false")
	}
	if ctx.IntMasterEnabled || ctx.enablingIme || ctx.imeEnableDelay != 0 {
		t.Fatalf("DI did not cancel pending EI: IME=%t enabling=%t delay=%d", ctx.IntMasterEnabled, ctx.enablingIme, ctx.imeEnableDelay)
	}
	if ctx.Regs.Pc != 0x0102 {
		t.Fatalf("PC after DI = %04X, want 0102", ctx.Regs.Pc)
	}

	if !ctx.Step() {
		t.Fatalf("NOP step returned false")
	}
	if ctx.Regs.Pc != 0x0103 {
		t.Fatalf("interrupt was serviced after EI;DI sequence, PC = %04X", ctx.Regs.Pc)
	}
	if ctx.IntFlags&byte(IT_TIMER) == 0 {
		t.Fatalf("pending timer interrupt was unexpectedly cleared")
	}
}

func TestHaltWakesOnlyForEnabledInterrupts(t *testing.T) {
	ctx := newInstructionTestCPU()
	ctx.Regs.Pc = 0x0100
	ctx.Halted = true
	ctx.IntFlags = byte(IT_VBLANK)

	bus := memory.BusCtx()
	bus.BusWrite(0xFFFF, byte(IT_TIMER))

	if !ctx.Step() {
		t.Fatalf("halted step returned false")
	}
	if !ctx.Halted {
		t.Fatalf("HALT woke for disabled VBlank interrupt")
	}
	if ctx.Regs.Pc != 0x0100 {
		t.Fatalf("PC changed while halted: %04X", ctx.Regs.Pc)
	}

	ctx.IntFlags |= byte(IT_TIMER)
	if !ctx.Step() {
		t.Fatalf("halt wake step returned false")
	}
	if ctx.Halted {
		t.Fatalf("HALT did not wake for enabled timer interrupt")
	}
	if ctx.Regs.Pc != 0x0100 {
		t.Fatalf("PC changed while waking from HALT without IME: %04X", ctx.Regs.Pc)
	}
	if ctx.IntFlags&byte(IT_TIMER) == 0 {
		t.Fatalf("timer interrupt was serviced even though IME is disabled")
	}
}

type gbOpcodeSet struct {
	Unprefixed map[string]gbOpcodeSpec `json:"unprefixed"`
	CBPrefixed map[string]gbOpcodeSpec `json:"cbprefixed"`
}

type gbOpcodeSpec struct {
	Mnemonic string          `json:"mnemonic"`
	Operands []gbOperandSpec `json:"operands"`
}

type gbOperandSpec struct {
	Name      string `json:"name"`
	Immediate bool   `json:"immediate"`
	Increment bool   `json:"increment"`
	Decrement bool   `json:"decrement"`
}

func loadGBDevOpcodes(t *testing.T) gbOpcodeSet {
	t.Helper()

	data, err := os.ReadFile("testdata/gb-opcodes.json")
	if err != nil {
		t.Fatalf("read gbdev opcode fixture: %v", err)
	}

	var opcodes gbOpcodeSet
	if err := json.Unmarshal(data, &opcodes); err != nil {
		t.Fatalf("parse gbdev opcode fixture: %v", err)
	}

	if len(opcodes.Unprefixed) != 0x100 {
		t.Fatalf("gbdev unprefixed opcode count = %d, want 256", len(opcodes.Unprefixed))
	}
	if len(opcodes.CBPrefixed) != 0x100 {
		t.Fatalf("gbdev CB-prefixed opcode count = %d, want 256", len(opcodes.CBPrefixed))
	}

	return opcodes
}

func loadGBDevBaseInstructions(t *testing.T) [0x100]Instruction {
	t.Helper()

	opcodes := loadGBDevOpcodes(t)
	var instructions [0x100]Instruction
	for opcode := 0; opcode <= 0xFF; opcode++ {
		key := fmt.Sprintf("0x%02X", opcode)
		spec, ok := opcodes.Unprefixed[key]
		if !ok {
			t.Fatalf("gbdev fixture missing unprefixed opcode %s", key)
		}

		instruction, err := instructionFromGBDevSpec(spec)
		if err != nil {
			t.Fatalf("map gbdev opcode %s (%s): %v", key, spec.Mnemonic, err)
		}
		instructions[opcode] = instruction
	}

	return instructions
}

func instructionFromGBDevSpec(spec gbOpcodeSpec) (Instruction, error) {
	if strings.HasPrefix(spec.Mnemonic, "ILLEGAL_") {
		return Instruction{}, nil
	}

	operands := spec.Operands
	condition := CT_NONE
	if gbDevMnemonicCanHaveCondition(spec.Mnemonic) && len(operands) > 0 {
		if parsed, ok := conditionFromGBDevName(operands[0].Name); ok {
			condition = parsed
			operands = operands[1:]
		}
	}

	switch spec.Mnemonic {
	case "NOP":
		return Instruction{Type: IN_NOP, Mode: AM_IMP}, nil
	case "LD":
		return ldInstructionFromGBDev(IN_LD, operands, condition)
	case "LDH":
		return ldInstructionFromGBDev(IN_LDH, operands, condition)
	case "INC":
		return registerOnlyInstructionFromGBDev(IN_INC, operands, condition)
	case "DEC":
		return registerOnlyInstructionFromGBDev(IN_DEC, operands, condition)
	case "RLCA":
		return Instruction{Type: IN_RLCA}, nil
	case "ADD":
		return aluInstructionFromGBDev(IN_ADD, operands, condition)
	case "RRCA":
		return Instruction{Type: IN_RRCA}, nil
	case "STOP":
		return Instruction{Type: IN_STOP, Mode: AM_D8}, nil
	case "RLA":
		return Instruction{Type: IN_RLA}, nil
	case "JR":
		return Instruction{Type: IN_JR, Mode: AM_D8, Condition: condition}, nil
	case "RRA":
		return Instruction{Type: IN_RRA}, nil
	case "DAA":
		return Instruction{Type: IN_DAA}, nil
	case "CPL":
		return Instruction{Type: IN_CPL}, nil
	case "SCF":
		return Instruction{Type: IN_SCF}, nil
	case "CCF":
		return Instruction{Type: IN_CCF}, nil
	case "HALT":
		return Instruction{Type: IN_HALT}, nil
	case "ADC":
		return aluInstructionFromGBDev(IN_ADC, operands, condition)
	case "SUB":
		return aluInstructionFromGBDev(IN_SUB, operands, condition)
	case "SBC":
		return aluInstructionFromGBDev(IN_SBC, operands, condition)
	case "AND":
		return aluInstructionFromGBDev(IN_AND, operands, condition)
	case "XOR":
		return aluInstructionFromGBDev(IN_XOR, operands, condition)
	case "OR":
		return aluInstructionFromGBDev(IN_OR, operands, condition)
	case "CP":
		return aluInstructionFromGBDev(IN_CP, operands, condition)
	case "POP":
		return registerOnlyInstructionFromGBDev(IN_POP, operands, condition)
	case "JP":
		return jumpInstructionFromGBDev(IN_JP, operands, condition)
	case "PUSH":
		return registerOnlyInstructionFromGBDev(IN_PUSH, operands, condition)
	case "RET":
		return Instruction{Type: IN_RET, Mode: AM_IMP, Condition: condition}, nil
	case "PREFIX":
		return Instruction{Type: IN_CB, Mode: AM_D8}, nil
	case "CALL":
		return Instruction{Type: IN_CALL, Mode: AM_D16, Condition: condition}, nil
	case "RETI":
		return Instruction{Type: IN_RETI}, nil
	case "DI":
		return Instruction{Type: IN_DI}, nil
	case "EI":
		return Instruction{Type: IN_EI}, nil
	case "RST":
		return rstInstructionFromGBDev(operands, condition)
	default:
		return Instruction{}, fmt.Errorf("unsupported mnemonic")
	}
}

func ldInstructionFromGBDev(inType InType, operands []gbOperandSpec, condition conditionTypes) (Instruction, error) {
	if condition != CT_NONE {
		return Instruction{}, fmt.Errorf("LD cannot be conditional")
	}

	if inType == IN_LDH && len(operands) == 2 {
		dst := operands[0]
		src := operands[1]
		if dst.Name == "a8" && !dst.Immediate && src.Name == "A" {
			return Instruction{Type: IN_LDH, Mode: AM_A8_R, Reg2: RT_A}, nil
		}
		if dst.Name == "A" && src.Name == "a8" && !src.Immediate {
			return Instruction{Type: IN_LDH, Mode: AM_R_A8, Reg1: RT_A}, nil
		}
		if dst.Name == "C" && !dst.Immediate && src.Name == "A" {
			return Instruction{Type: IN_LD, Mode: AM_MR_R, Reg1: RT_C, Reg2: RT_A}, nil
		}
		if dst.Name == "A" && src.Name == "C" && !src.Immediate {
			return Instruction{Type: IN_LD, Mode: AM_R_MR, Reg1: RT_A, Reg2: RT_C}, nil
		}
	}

	if len(operands) == 3 &&
		operands[0].Name == "HL" &&
		operands[1].Name == "SP" &&
		operands[2].Name == "e8" {
		return Instruction{Type: IN_LD, Mode: AM_HL_SPR, Reg1: RT_HL, Reg2: RT_SP}, nil
	}

	if len(operands) != 2 {
		return Instruction{}, fmt.Errorf("LD operand count = %d, want 2", len(operands))
	}

	dst := operands[0]
	src := operands[1]
	if dst.Name == "a16" && !dst.Immediate {
		srcReg, ok := regFromGBDevName(src.Name)
		if !ok {
			return Instruction{}, fmt.Errorf("LD [a16] source %q is not a register", src.Name)
		}
		return Instruction{Type: IN_LD, Mode: AM_A16_R, Reg2: srcReg}, nil
	}
	if dst.Name == "A" && src.Name == "a16" && !src.Immediate {
		return Instruction{Type: IN_LD, Mode: AM_R_A16, Reg1: RT_A}, nil
	}

	if dst.Name == "HL" && dst.Increment && !dst.Immediate {
		return Instruction{Type: IN_LD, Mode: AM_HLI_R, Reg1: RT_HL, Reg2: RT_A}, nil
	}
	if dst.Name == "HL" && dst.Decrement && !dst.Immediate {
		return Instruction{Type: IN_LD, Mode: AM_HLD_R, Reg1: RT_HL, Reg2: RT_A}, nil
	}
	if dst.Name == "A" && src.Name == "HL" && src.Increment && !src.Immediate {
		return Instruction{Type: IN_LD, Mode: AM_R_HLI, Reg1: RT_A, Reg2: RT_HL}, nil
	}
	if dst.Name == "A" && src.Name == "HL" && src.Decrement && !src.Immediate {
		return Instruction{Type: IN_LD, Mode: AM_R_HLD, Reg1: RT_A, Reg2: RT_HL}, nil
	}

	dstReg, dstIsReg := regFromGBDevName(dst.Name)
	if !dstIsReg {
		return Instruction{}, fmt.Errorf("LD destination %q is not supported", dst.Name)
	}

	if !dst.Immediate {
		if src.Name == "n8" {
			return Instruction{Type: IN_LD, Mode: AM_MR_D8, Reg1: dstReg}, nil
		}
		srcReg, ok := regFromGBDevName(src.Name)
		if !ok {
			return Instruction{}, fmt.Errorf("LD memory source %q is not a register", src.Name)
		}
		return Instruction{Type: IN_LD, Mode: AM_MR_R, Reg1: dstReg, Reg2: srcReg}, nil
	}

	if !src.Immediate {
		srcReg, ok := regFromGBDevName(src.Name)
		if !ok {
			return Instruction{}, fmt.Errorf("LD source %q is not a register", src.Name)
		}
		return Instruction{Type: IN_LD, Mode: AM_R_MR, Reg1: dstReg, Reg2: srcReg}, nil
	}

	switch src.Name {
	case "n8":
		return Instruction{Type: IN_LD, Mode: AM_R_D8, Reg1: dstReg}, nil
	case "n16":
		return Instruction{Type: IN_LD, Mode: AM_R_D16, Reg1: dstReg}, nil
	default:
		srcReg, ok := regFromGBDevName(src.Name)
		if !ok {
			return Instruction{}, fmt.Errorf("LD source %q is not supported", src.Name)
		}
		return Instruction{Type: IN_LD, Mode: AM_R_R, Reg1: dstReg, Reg2: srcReg}, nil
	}
}

func registerOnlyInstructionFromGBDev(inType InType, operands []gbOperandSpec, condition conditionTypes) (Instruction, error) {
	if condition != CT_NONE {
		return Instruction{}, fmt.Errorf("%s cannot be conditional", getInstructionName(inType))
	}
	if len(operands) != 1 {
		return Instruction{}, fmt.Errorf("%s operand count = %d, want 1", getInstructionName(inType), len(operands))
	}

	reg, ok := regFromGBDevName(operands[0].Name)
	if !ok {
		return Instruction{}, fmt.Errorf("operand %q is not a register", operands[0].Name)
	}
	mode := AM_R
	if !operands[0].Immediate {
		mode = AM_MR
	}

	return Instruction{Type: inType, Mode: mode, Reg1: reg}, nil
}

func aluInstructionFromGBDev(inType InType, operands []gbOperandSpec, condition conditionTypes) (Instruction, error) {
	if condition != CT_NONE {
		return Instruction{}, fmt.Errorf("%s cannot be conditional", getInstructionName(inType))
	}
	if len(operands) != 2 {
		return Instruction{}, fmt.Errorf("%s operand count = %d, want 2", getInstructionName(inType), len(operands))
	}

	dst := operands[0]
	src := operands[1]
	if inType == IN_ADD && dst.Name == "HL" {
		srcReg, ok := regFromGBDevName(src.Name)
		if !ok {
			return Instruction{}, fmt.Errorf("ADD HL source %q is not a register", src.Name)
		}
		return Instruction{Type: IN_ADD, Mode: AM_R_R, Reg1: RT_HL, Reg2: srcReg}, nil
	}
	if inType == IN_ADD && dst.Name == "SP" && src.Name == "e8" {
		return Instruction{Type: IN_ADD, Mode: AM_R_D8, Reg1: RT_SP}, nil
	}
	if dst.Name != "A" {
		return Instruction{}, fmt.Errorf("%s destination %q is not A", getInstructionName(inType), dst.Name)
	}

	if src.Name == "n8" {
		return Instruction{Type: inType, Mode: AM_R_D8, Reg1: RT_A}, nil
	}

	srcReg, ok := regFromGBDevName(src.Name)
	if !ok {
		return Instruction{}, fmt.Errorf("%s source %q is not supported", getInstructionName(inType), src.Name)
	}
	mode := AM_R_R
	if !src.Immediate {
		mode = AM_R_MR
	}

	return Instruction{Type: inType, Mode: mode, Reg1: RT_A, Reg2: srcReg}, nil
}

func jumpInstructionFromGBDev(inType InType, operands []gbOperandSpec, condition conditionTypes) (Instruction, error) {
	if len(operands) != 1 {
		return Instruction{}, fmt.Errorf("%s operand count = %d, want 1", getInstructionName(inType), len(operands))
	}

	if operands[0].Name == "HL" {
		return Instruction{Type: inType, Mode: AM_R, Reg1: RT_HL, Condition: condition}, nil
	}
	if operands[0].Name == "a16" {
		return Instruction{Type: inType, Mode: AM_D16, Condition: condition}, nil
	}

	return Instruction{}, fmt.Errorf("%s operand %q is not supported", getInstructionName(inType), operands[0].Name)
}

func rstInstructionFromGBDev(operands []gbOperandSpec, condition conditionTypes) (Instruction, error) {
	if condition != CT_NONE {
		return Instruction{}, fmt.Errorf("RST cannot be conditional")
	}
	if len(operands) != 1 {
		return Instruction{}, fmt.Errorf("RST operand count = %d, want 1", len(operands))
	}

	param, err := strconv.ParseUint(strings.TrimPrefix(operands[0].Name, "$"), 16, 8)
	if err != nil {
		return Instruction{}, fmt.Errorf("parse RST target %q: %w", operands[0].Name, err)
	}

	return Instruction{Type: IN_RST, Mode: AM_IMP, Param: byte(param)}, nil
}

func regFromGBDevName(name string) (regTypes, bool) {
	switch name {
	case "A":
		return RT_A, true
	case "F":
		return RT_F, true
	case "B":
		return RT_B, true
	case "C":
		return RT_C, true
	case "D":
		return RT_D, true
	case "E":
		return RT_E, true
	case "H":
		return RT_H, true
	case "L":
		return RT_L, true
	case "AF":
		return RT_AF, true
	case "BC":
		return RT_BC, true
	case "DE":
		return RT_DE, true
	case "HL":
		return RT_HL, true
	case "SP":
		return RT_SP, true
	case "PC":
		return RT_PC, true
	default:
		return RT_NONE, false
	}
}

func conditionFromGBDevName(name string) (conditionTypes, bool) {
	switch name {
	case "NZ":
		return CT_NZ, true
	case "Z":
		return CT_Z, true
	case "NC":
		return CT_NC, true
	case "C":
		return CT_C, true
	default:
		return CT_NONE, false
	}
}

func gbDevMnemonicCanHaveCondition(mnemonic string) bool {
	switch mnemonic {
	case "JR", "JP", "CALL", "RET":
		return true
	default:
		return false
	}
}

func assertGBDevCBSpec(t *testing.T, spec gbOpcodeSpec, opcode byte) {
	t.Helper()

	regNames := []string{"B", "C", "D", "E", "H", "L", "HL", "A"}
	targetReg := regNames[opcode&0x07]
	bit := (opcode >> 3) & 0x07

	switch opcode >> 6 {
	case 0:
		mnemonics := []string{"RLC", "RRC", "RL", "RR", "SLA", "SRA", "SWAP", "SRL"}
		assertGBDevMnemonic(t, spec, mnemonics[bit])
		assertGBDevOperands(t, spec, targetReg)
	case 1:
		assertGBDevMnemonic(t, spec, "BIT")
		assertGBDevOperands(t, spec, fmt.Sprintf("%d", bit), targetReg)
	case 2:
		assertGBDevMnemonic(t, spec, "RES")
		assertGBDevOperands(t, spec, fmt.Sprintf("%d", bit), targetReg)
	case 3:
		assertGBDevMnemonic(t, spec, "SET")
		assertGBDevOperands(t, spec, fmt.Sprintf("%d", bit), targetReg)
	}
}

func assertGBDevMnemonic(t *testing.T, spec gbOpcodeSpec, mnemonic string) {
	t.Helper()

	if spec.Mnemonic != mnemonic {
		t.Fatalf("gbdev mnemonic = %q, want %q", spec.Mnemonic, mnemonic)
	}
}

func assertGBDevOperands(t *testing.T, spec gbOpcodeSpec, operands ...string) {
	t.Helper()

	if len(spec.Operands) != len(operands) {
		t.Fatalf("gbdev operand count = %d, want %d", len(spec.Operands), len(operands))
	}
	for i, operand := range operands {
		if spec.Operands[i].Name != operand {
			t.Fatalf("gbdev operand %d = %q, want %q", i, spec.Operands[i].Name, operand)
		}
	}
}

func cbInitialValue(opcode byte) byte {
	switch opcode >> 6 {
	case 0:
		switch (opcode >> 3) & 0x07 {
		case 0:
			return 0x80
		case 1:
			return 0x01
		case 2:
			return 0x80
		case 3:
			return 0x01
		case 4:
			return 0x80
		case 5:
			return 0x81
		case 6:
			return 0xF0
		case 7:
			return 0x01
		}
	case 1:
		return 0x55
	case 2:
		return 0xFF
	case 3:
		return 0x00
	}

	return 0x00
}

func expectedCBResult(opcode byte, value byte, carryIn bool) (byte, byte) {
	bit := (opcode >> 3) & 0x07

	switch opcode >> 6 {
	case 0:
		switch bit {
		case 0:
			result := (value << 1) | (value >> 7)
			return result, flags(result == 0, false, false, value&0x80 != 0)
		case 1:
			result := (value >> 1) | (value << 7)
			return result, flags(result == 0, false, false, value&0x01 != 0)
		case 2:
			result := value << 1
			if carryIn {
				result |= 0x01
			}
			return result, flags(result == 0, false, false, value&0x80 != 0)
		case 3:
			result := value >> 1
			if carryIn {
				result |= 0x80
			}
			return result, flags(result == 0, false, false, value&0x01 != 0)
		case 4:
			result := value << 1
			return result, flags(result == 0, false, false, value&0x80 != 0)
		case 5:
			result := (value >> 1) | (value & 0x80)
			return result, flags(result == 0, false, false, value&0x01 != 0)
		case 6:
			result := (value >> 4) | (value << 4)
			return result, flags(result == 0, false, false, false)
		case 7:
			result := value >> 1
			return result, flags(result == 0, false, false, value&0x01 != 0)
		}
	case 1:
		return value, flags(value&(1<<bit) == 0, false, true, carryIn)
	case 2:
		return value &^ (1 << bit), flags(false, false, false, carryIn)
	case 3:
		return value | (1 << bit), flags(false, false, false, carryIn)
	}

	return value, flags(false, false, false, carryIn)
}

func flags(z bool, n bool, h bool, c bool) byte {
	var out byte
	if z {
		out |= 0x80
	}
	if n {
		out |= 0x40
	}
	if h {
		out |= 0x20
	}
	if c {
		out |= 0x10
	}
	return out
}

func newInstructionTestCPU() *CpuContext {
	mem := &testMemory{}
	bus := memory.NewBus(mem, mem, mem, mem, mem, nil)
	ctx := NewCpuContext(bus)
	CpuSetReg(RT_HL, 0xC123)
	Cm.ticks = 0
	timerInstance = &TimerContext{div: 0xAC00}
	return ctx
}

func seedStepState(ctx *CpuContext, opcode byte) {
	ctx.Regs = CpuRegisters{
		A:  0x42,
		F:  0x10,
		B:  0x12,
		C:  0x34,
		D:  0x56,
		E:  0x78,
		H:  0xC1,
		L:  0x23,
		Pc: 0x0100,
		Sp: 0xD000,
	}

	bus := memory.BusCtx()
	bus.BusWrite(0x0100, opcode)
	bus.BusWrite(0x0101, 0x34)
	bus.BusWrite(0x0102, 0x12)
	bus.BusWrite(0x1234, 0xBC)
	bus.BusWrite(0xC123, 0x5A)
	bus.BusWrite(0xD000, 0x78)
	bus.BusWrite(0xD001, 0x56)
	bus.BusWrite(0xFF34, 0x9A)
}

func writeCBTarget(reg regTypes, value byte) {
	if reg == RT_HL {
		memory.BusCtx().BusWrite(CpuRegRead(RT_HL), value)
		return
	}

	CpuSetReg8(reg, value)
}

func readCBTarget(reg regTypes) byte {
	if reg == RT_HL {
		return memory.BusCtx().BusRead(CpuRegRead(RT_HL))
	}

	return CpuRegRead8(reg)
}

type testMemory struct {
	data [0x10000]byte
}

func (m *testMemory) CartRead(address uint16) byte {
	return m.data[address]
}

func (m *testMemory) CartWrite(address uint16, data byte) {
	m.data[address] = data
}

func (m *testMemory) WramRead(address uint16) byte {
	return m.data[address]
}

func (m *testMemory) WramWrite(address uint16, value byte) {
	m.data[address] = value
}

func (m *testMemory) HramRead(address uint16) byte {
	return m.data[address]
}

func (m *testMemory) HramWrite(address uint16, value byte) {
	m.data[address] = value
}

func (m *testMemory) DMATransferring() bool {
	return false
}

func (m *testMemory) OamRead(address uint16) byte {
	return m.data[address]
}

func (m *testMemory) OamWrite(address uint16, value byte) {
	m.data[address] = value
}

func (m *testMemory) VramRead(address uint16) byte {
	return m.data[address]
}

func (m *testMemory) VramWrite(address uint16, value byte) {
	m.data[address] = value
}

func (m *testMemory) Read(address uint16) byte {
	return m.data[address]
}

func (m *testMemory) Write(address uint16, value byte) {
	m.data[address] = value
}
