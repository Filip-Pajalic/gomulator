package cpu

import (
	"pajalic.go.emulator/packages/pubsub"
)

func StackPush(data byte) {
	regs := CpuGetRegs()
	regs.Sp--
	pubsub.BusCtx().BusWrite(regs.Sp, data)
}

func StackPush16(data uint16) {
	lowByte := byte(data & 0xFF)
	highByte := byte((data >> 8) & 0xFF)
	StackPush(highByte)
	StackPush(lowByte)
}

func StackPop() byte {
	regs := CpuGetRegs()
	result := pubsub.BusCtx().BusRead(regs.Sp)
	regs.Sp++
	return result
}

func StackPop16() uint16 {
	low := uint16(StackPop())
	high := uint16(StackPop())
	return (high << 8) | low
}
