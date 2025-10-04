package tests

import (
	"testing"

	gameboypackage "app/internal/cpu"
)

func TestNothing(t *testing.T) {
	got := gameboypackage.CpuStep()
	want := false
	if got != want {
		t.Fatalf("got %T, wanted %T", got, want)
	}
}
