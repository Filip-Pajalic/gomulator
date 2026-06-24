package input

import (
	"app/internal/cpu"
	"testing"
)

func newJoypadTestCPU() *cpu.CpuContext {
	Init()
	ctx := cpu.NewCpuContext(nil)
	ctx.IntFlags = 0
	return ctx
}

func TestSetStateRequestsJoypadInterruptForSelectedButtonPress(t *testing.T) {
	ctx := newJoypadTestCPU()
	SetSel(0x10) // Button group selected, direction group not selected.

	SetState(State{Start: true})

	if ctx.IntFlags&byte(cpu.IT_JOYPAD) == 0 {
		t.Fatalf("expected JOYPAD interrupt, IF=%02X", ctx.IntFlags)
	}
}

func TestSetStateDoesNotRequestJoypadInterruptForUnselectedButton(t *testing.T) {
	ctx := newJoypadTestCPU()
	SetSel(0x30) // Neither group selected.

	SetState(State{Start: true})

	if ctx.IntFlags&byte(cpu.IT_JOYPAD) != 0 {
		t.Fatalf("unexpected JOYPAD interrupt, IF=%02X", ctx.IntFlags)
	}
}

func TestSetSelRequestsJoypadInterruptWhenHeldButtonBecomesSelected(t *testing.T) {
	ctx := newJoypadTestCPU()
	SetState(State{Start: true})
	ctx.IntFlags = 0

	SetSel(0x10) // Selecting the held button line drives P1 bit 3 low.

	if ctx.IntFlags&byte(cpu.IT_JOYPAD) == 0 {
		t.Fatalf("expected JOYPAD interrupt, IF=%02X", ctx.IntFlags)
	}
}

func TestSetStateDoesNotRequestJoypadInterruptOnReleaseOrRepeat(t *testing.T) {
	ctx := newJoypadTestCPU()
	SetSel(0x10)
	SetState(State{Start: true})
	ctx.IntFlags = 0

	SetState(State{Start: true})
	if ctx.IntFlags&byte(cpu.IT_JOYPAD) != 0 {
		t.Fatalf("unexpected repeat JOYPAD interrupt, IF=%02X", ctx.IntFlags)
	}

	SetState(State{})
	if ctx.IntFlags&byte(cpu.IT_JOYPAD) != 0 {
		t.Fatalf("unexpected release JOYPAD interrupt, IF=%02X", ctx.IntFlags)
	}
}
