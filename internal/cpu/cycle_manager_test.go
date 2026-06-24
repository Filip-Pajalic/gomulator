package cpu

import "testing"

func TestCycleManagerTicksPeripheralHookInTCycles(t *testing.T) {
	cm := &CycleManager{}
	var got int32
	cm.SetPeripheralTick(func(tCycles int32) {
		got += tCycles
	})

	cm.IncreaseCycle(3)

	if got != 12 {
		t.Fatalf("peripheral ticks = %d, want 12", got)
	}
}

func TestNewCpuContextClearsPeripheralHook(t *testing.T) {
	called := false
	Cm.SetPeripheralTick(func(int32) {
		called = true
	})

	NewCpuContext(nil)
	Cm.IncreaseCycle(1)

	if called {
		t.Fatal("peripheral hook still installed after NewCpuContext")
	}
}
