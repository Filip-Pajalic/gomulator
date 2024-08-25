package cpu

import (
	"pajalic.go.emulator/packages/pubsub"
)

func StackPush(data byte) {
	CpuGetRegs().Sp--
	pubsub.BusCtx().BusWrite(CpuGetRegs().Sp, data)

}

func StackPush16(data uint16) {
	value := (data >> 8) & 0xFF
	value2 := data & 0xFF
	StackPush(byte(value))
	StackPush(byte(value2))
}

func StackPop() byte {
	result := pubsub.BusCtx().BusRead(CpuGetRegs().Sp)
	CpuGetRegs().Sp++
	return result
}

func StackPop16() uint16 {
	var lo = uint16(StackPop())
	var hi = uint16(StackPop())

	return (hi << 8) | lo
}
