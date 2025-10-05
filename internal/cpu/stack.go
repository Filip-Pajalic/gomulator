package cpu

import (
	logger "app/internal/logger"
	"app/internal/memory"
)

// StackPush: Pushes a byte onto the stack (decrement SP, then write)
func StackPush(data byte) {
	regs := CpuGetRegs()
	regs.Sp--
	addr := regs.Sp
	memory.BusCtx().BusWrite(addr, data)
	logger.Debug("StackPush: wrote %02X to %04X (SP now %04X)", data, addr, regs.Sp)
}

// StackPush16: Pushes a 16-bit value onto the stack (high byte first, then low byte)
// Game Boy is little-endian, so low byte is at lower address
func StackPush16(data uint16) {
	// Push high byte, then low byte (so low byte is at SP, high byte at SP-1)
	highByte := byte((data >> 8) & 0xFF)
	lowByte := byte(data & 0xFF)
	StackPush(highByte)
	StackPush(lowByte)
}

// StackPop: Pops a byte from the stack (read, then increment SP)
func StackPop() byte {
	regs := CpuGetRegs()
	result := memory.BusCtx().BusRead(regs.Sp)
	regs.Sp++
	return result
}

// StackPop16: Pops a 16-bit value from the stack (low byte first, then high byte)
func StackPop16() uint16 {
	low := uint16(StackPop())
	high := uint16(StackPop())
	return (high << 8) | low
}
