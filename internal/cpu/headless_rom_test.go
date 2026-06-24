package cpu

import (
	"app/internal/memory"
	"fmt"
	"os"
	"testing"
)

type headlessPPU struct {
	vram [0x2000]byte
	oam  [0xA0]byte
	ly   byte
}

func (p *headlessPPU) VramRead(address uint16) byte {
	if address < 0x8000 || address >= 0xA000 {
		return 0xFF
	}
	return p.vram[address-0x8000]
}

func (p *headlessPPU) VramWrite(address uint16, value byte) {
	if address >= 0x8000 && address < 0xA000 {
		p.vram[address-0x8000] = value
	}
}

func (p *headlessPPU) OamRead(address uint16) byte {
	if address < 0xFE00 || address >= 0xFEA0 {
		return 0xFF
	}
	return p.oam[address-0xFE00]
}

func (p *headlessPPU) OamWrite(address uint16, value byte) {
	if address >= 0xFE00 && address < 0xFEA0 {
		p.oam[address-0xFE00] = value
	}
}

type headlessIO struct {
	timer *TimerContext
	dma   *DMAContext
	ppu   *headlessPPU
	lcd   [0x0C]byte
	sb    byte
	sc    byte
	key1  byte
}

func (i *headlessIO) Read(address uint16) byte {
	switch {
	case address == 0xFF00:
		return 0xCF
	case address == 0xFF01:
		return i.sb
	case address == 0xFF02:
		return i.sc
	case address >= 0xFF04 && address <= 0xFF07:
		return i.timer.Read(address)
	case address == 0xFF0F:
		return CpuGetIntFlags()
	case address == 0xFF44:
		i.ppu.ly++
		return i.ppu.ly
	case address == 0xFF4D:
		return i.key1
	case address >= 0xFF40 && address <= 0xFF4B:
		return i.lcd[address-0xFF40]
	case address == 0xFF70:
		return memory.ReadWramBank()
	default:
		return 0
	}
}

func (i *headlessIO) Write(address uint16, value byte) {
	switch {
	case address == 0xFF01:
		i.sb = value
	case address == 0xFF02:
		i.sc = value
	case address >= 0xFF04 && address <= 0xFF07:
		i.timer.Write(address, value)
	case address == 0xFF0F:
		CpuSetIntFlags(value)
	case address == 0xFF44:
		i.ppu.ly = value
	case address == 0xFF46:
		i.dma.RestartDMAContext(value)
	case address == 0xFF4D:
		i.key1 = (i.key1 & 0x80) | (value & 0x01)
	case address >= 0xFF40 && address <= 0xFF4B:
		i.lcd[address-0xFF40] = value
	case address == 0xFF70:
		memory.WriteWramBank(value)
	}
}

func (i *headlessIO) TrySpeedSwitch() bool {
	if i.key1&0x01 == 0 {
		return false
	}

	i.key1 = (i.key1 ^ 0x80) & 0x80
	return true
}

func TestHeadlessDebugROM(t *testing.T) {
	rom := os.Getenv("GOMULATOR_HEADLESS_ROM")
	if rom == "" {
		t.Skip("set GOMULATOR_HEADLESS_ROM to run a ROM headlessly")
	}

	Cm.ticks = 0
	ResetDebugTestResult()

	cart := memory.CartCtx()
	if !cart.CartLoad(rom) {
		t.Fatalf("failed to load ROM %q", rom)
	}

	timer := TimerCtx()
	dma := DmaCtx()
	ppu := &headlessPPU{}
	ram := memory.RamCtx()
	io := &headlessIO{timer: timer, dma: dma, ppu: ppu}

	cpu := NewCpuContext(nil)
	bus := memory.NewBus(cart, ram, dma, ppu, io, cpu)
	cpu = NewCpuContext(bus)

	const maxTicks int32 = 650_000_000
	for cpu.Step() && Cm.GetCycleTicks() < maxTicks {
		dma.DMATickBatch(4)
		switch GetDebugTestResult() {
		case DebugTestPassed:
			return
		case DebugTestFailed:
			t.Fatalf("debug ROM failed at ticks=%d output=%q", Cm.GetCycleTicks(), string(dbgMsg[:msgSize]))
		}
	}
	if GetDebugTestResult() == DebugTestPassed {
		return
	}

	t.Fatalf("debug ROM did not finish: ticks=%d result=%d halted=%t pc=%04X sp=%04X output=%q code=%s stack=%s bss=%s",
		Cm.GetCycleTicks(),
		GetDebugTestResult(),
		cpu.Halted,
		cpu.Regs.Pc,
		cpu.Regs.Sp,
		string(dbgMsg[:msgSize]),
		dumpBus(cpu.Regs.Pc-16, 48),
		dumpBus(cpu.Regs.Sp, 16),
		dumpBus(0xD800, 16),
	)
}

func dumpBus(start uint16, n int) string {
	bus := memory.BusCtx()
	if bus == nil {
		return ""
	}
	out := make([]byte, 0, n*3)
	for i := 0; i < n; i++ {
		out = fmt.Appendf(out, "%02X", bus.BusRead(start+uint16(i)))
		if i+1 < n {
			out = append(out, ' ')
		}
	}
	return string(out)
}
