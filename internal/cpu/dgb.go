package cpu

import (
	"app/internal/memory"
	"strings"

	logger "app/internal/logger"
)

const (
	nextChecksumAddr uint16 = 0xD800
	resultAddr       uint16 = 0xD802
	testNameAddr     uint16 = 0xD803
	tempAddr         uint16 = 0xD805
)

var dbgMsg [1024]byte
var msgSize = 0
var nextDebugProgressTick int32 = debugProgressInterval

const debugProgressInterval int32 = 20_000_000

type DebugTestResult byte

const (
	DebugTestNone DebugTestResult = iota
	DebugTestPassed
	DebugTestFailed
)

func ResetDebugTestResult() {
	msgSize = 0
	nextDebugProgressTick = debugProgressInterval
	debugTestResult = DebugTestNone
}

func GetDebugTestResult() DebugTestResult {
	return debugTestResult
}

var debugTestResult = DebugTestNone

func DbgUpdate() {
	if memory.BusCtx().BusRead(0xFF02) == 0x81 {
		var c = memory.BusCtx().BusRead(0xFF01)

		if msgSize < len(dbgMsg) {
			dbgMsg[msgSize] = c
			msgSize++

			// Log character received (for debugging)
			if c >= 32 && c <= 126 { // printable ASCII
				logger.Debug("Serial received: '%c' (0x%02X)", c, c)
			} else {
				logger.Debug("Serial received: 0x%02X", c)
			}
		} else {
			logger.Warn("dbgMsg buffer overflow")
			msgSize = 0 // Reset to avoid further errors
		}

		memory.BusCtx().BusWrite(0xFF02, 0)
	}
}

func DbgProgress() {
	ticks := Cm.GetCycleTicks()
	if ticks < nextDebugProgressTick {
		return
	}
	for ticks >= nextDebugProgressTick {
		nextDebugProgressTick += debugProgressInterval
	}

	if cpuInstance == nil || memory.BusCtx() == nil {
		return
	}

	bus := memory.BusCtx()
	regs := cpuInstance.Regs
	logger.Info("DEBUG PROGRESS ticks=%d PC=%04X SP=%04X OP=%02X AF=%02X%02X BC=%02X%02X DE=%02X%02X HL=%02X%02X HALT=%t STOP=%t ROMBANK=%d IF=%02X IE=%02X LY=%02X SB=%02X SC=%02X MSG=%d",
		ticks,
		regs.Pc,
		regs.Sp,
		bus.BusRead(regs.Pc),
		regs.A,
		regs.F,
		regs.B,
		regs.C,
		regs.D,
		regs.E,
		regs.H,
		regs.L,
		cpuInstance.Halted,
		cpuInstance.Stopped,
		memory.CartCtx().CurrentROMBank(),
		cpuInstance.IntFlags,
		bus.GetInterruptEnable(),
		bus.BusRead(0xFF44),
		bus.BusRead(0xFF01),
		bus.BusRead(0xFF02),
		msgSize,
	)
}

func DbgPrint() bool {
	if msgSize > 0 {
		debugmsg := strings.TrimSpace(string(dbgMsg[:msgSize]))
		debugmsgLower := strings.ToLower(debugmsg)
		hasCompleteLine := dbgMsg[msgSize-1] == '\n'
		hasResultToken := strings.Contains(debugmsgLower, "passed") ||
			strings.Contains(debugmsgLower, "failed") ||
			strings.Contains(debugmsgLower, "error")

		// Check complete lines and terminal pass/fail messages. The Blargg
		// multi-ROM can leave the final result in the serial buffer without a
		// trailing newline, so don't wait forever once the result text appears.
		if hasCompleteLine || hasResultToken {
			if len(debugmsg) == 0 {
				logger.Debug("TEST OUTPUT RAW: % X", dbgMsg[:msgSize])

				// When tests emit a blank line, capture additional context to
				// identify the currently executing instruction and result code.
				if bus := memory.BusCtx(); bus != nil {
					// BSS layout from testing.s/checksums.s:
					//   0xD800-0xD801: next_checksum pointer
					//   0xD802:        result code
					//   0xD803-0xD804: test_name pointer
					result := bus.BusRead(resultAddr)
					instAddr := uint16(0xDEF8)
					op0 := bus.BusRead(instAddr)
					op1 := bus.BusRead(instAddr + 1)
					op2 := bus.BusRead(instAddr + 2)

					pc := uint16(0)
					sp := uint16(0)
					if cpuInstance != nil {
						pc = cpuInstance.Regs.Pc
						sp = cpuInstance.Regs.Sp
					}

					origSP := uint16(bus.BusRead(tempAddr)) | (uint16(bus.BusRead(tempAddr+1)) << 8)
					finalSP := uint16(bus.BusRead(tempAddr+2)) | (uint16(bus.BusRead(tempAddr+3)) << 8)
					testPtr := uint16(bus.BusRead(testNameAddr)) | (uint16(bus.BusRead(testNameAddr+1)) << 8)
					bssBytes := [10]byte{}
					for i := 0; i < len(bssBytes); i++ {
						bssBytes[i] = bus.BusRead(nextChecksumAddr + uint16(i))
					}

					logger.Debug("TEST CONTEXT: result=%02X instr=%02X %02X %02X PC=%04X SP=%04X SP_ORIG=%04X SP_FINAL=%04X TEST_PTR=%04X BSS=% X", result, op0, op1, op2, pc, sp, origSP, finalSP, testPtr, bssBytes)

					checksum := [4]byte{}
					for i := 0; i < 4; i++ {
						checksum[i] = bus.BusRead(0xFF80 + uint16(i))
					}
					nextPtrRaw := uint16(bus.BusRead(nextChecksumAddr)) | (uint16(bus.BusRead(nextChecksumAddr+1)) << 8)
					nextPtr := nextPtrRaw
					if nextPtr >= 4 {
						nextPtr -= 4
					}
					expected := [4]byte{}
					for i := 0; i < 4; i++ {
						expected[i] = bus.BusRead(nextPtr + uint16(i))
					}
					logger.Debug("TEST CRC: actual=%02X%02X%02X%02X expected=%02X%02X%02X%02X ptr=%04X raw=%04X", checksum[0], checksum[1], checksum[2], checksum[3], expected[0], expected[1], expected[2], expected[3], nextPtr, nextPtrRaw)

					crcSample := [16]byte{}
					for i := 0; i < len(crcSample); i++ {
						crcSample[i] = bus.BusRead(0xD900 + uint16(i))
					}
					logger.Debug("CRC table sample D900: % X", crcSample)
				}
			}
			logger.Debug("TEST OUTPUT: %s", debugmsg)

			msgSize = 0 // Reset msgSize after printing

			// Check for common test failure indicators
			if strings.Contains(debugmsgLower, "failed") || strings.Contains(debugmsgLower, "error") {
				if cpuInstance != nil {
					regs := cpuInstance.Regs
					sp := regs.Sp
					low := memory.BusCtx().BusRead(sp - 2)
					high := memory.BusCtx().BusRead(sp - 1)
					logger.Info("CPU STATE -> PC:%04X SP:%04X AF:%02X%02X BC:%02X%02X DE:%02X%02X HL:%02X%02X IF:%02X IE:%02X IME:%t EI_DEFER:%t STACK[%04X]=%02X STACK[%04X]=%02X",
						regs.Pc, regs.Sp,
						regs.A, regs.F,
						regs.B, regs.C,
						regs.D, regs.E,
						regs.H, regs.L,
						cpuInstance.IntFlags,
						memory.BusCtx().GetInterruptEnable(),
						cpuInstance.IntMasterEnabled,
						cpuInstance.enablingIme,
						sp-2, low,
						sp-1, high,
					)
				}

				if bus := memory.BusCtx(); bus != nil {
					result := bus.BusRead(resultAddr)
					instAddr := uint16(0xDEF8)
					op0 := bus.BusRead(instAddr)
					op1 := bus.BusRead(instAddr + 1)
					op2 := bus.BusRead(instAddr + 2)

					pc := uint16(0)
					sp := uint16(0)
					if cpuInstance != nil {
						pc = cpuInstance.Regs.Pc
						sp = cpuInstance.Regs.Sp
					}

					origSP := uint16(bus.BusRead(tempAddr)) | (uint16(bus.BusRead(tempAddr+1)) << 8)
					finalSP := uint16(bus.BusRead(tempAddr+2)) | (uint16(bus.BusRead(tempAddr+3)) << 8)
					testPtr := uint16(bus.BusRead(testNameAddr)) | (uint16(bus.BusRead(testNameAddr+1)) << 8)
					bssBytes := [10]byte{}
					for i := 0; i < len(bssBytes); i++ {
						bssBytes[i] = bus.BusRead(nextChecksumAddr + uint16(i))
					}

					logger.Info("TEST FAILURE CONTEXT: result=%02X instr=%02X %02X %02X PC=%04X SP=%04X SP_ORIG=%04X SP_FINAL=%04X TEST_PTR=%04X BSS=% X", result, op0, op1, op2, pc, sp, origSP, finalSP, testPtr, bssBytes)

					hram := [4]byte{}
					for i := 0; i < len(hram); i++ {
						hram[i] = bus.BusRead(0xFF80 + uint16(i))
					}

					nextPtrRaw := uint16(bus.BusRead(nextChecksumAddr)) | (uint16(bus.BusRead(nextChecksumAddr+1)) << 8)
					nextPtr := nextPtrRaw
					if nextPtr >= 4 {
						nextPtr -= 4
					}
					expected := [4]byte{}
					for i := 0; i < len(expected); i++ {
						expected[i] = bus.BusRead(nextPtr + uint16(i))
					}
					logger.Info("TEST FAILURE CRC: actual=%02X%02X%02X%02X expected=%02X%02X%02X%02X ptr=%04X raw=%04X", hram[0], hram[1], hram[2], hram[3], expected[0], expected[1], expected[2], expected[3], nextPtr, nextPtrRaw)
					crcSample := [16]byte{}
					for i := 0; i < len(crcSample); i++ {
						crcSample[i] = bus.BusRead(0xD900 + uint16(i))
					}
					logger.Debug("CRC table sample D900: % X", crcSample)
				}
				logger.Info("*** TEST FAILED ***")
				debugTestResult = DebugTestFailed
				return false
			}

			// Check for success indicators
			if strings.Contains(debugmsgLower, "passed") {
				logger.Info("*** TEST PASSED ***")
				debugTestResult = DebugTestPassed
				return false
			}
		}
	}
	return true
}
