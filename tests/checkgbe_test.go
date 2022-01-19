package tests

import (
	gameboypackage "pajalic.go.emulator/packages"
	"testing"
)

func TestNothing(t *testing.T) {
	got := gameboypackage.CpuStep()
	want := false
	if got != want {
		t.Fatalf("got %T, wanted %T", got, want)
	}
}
