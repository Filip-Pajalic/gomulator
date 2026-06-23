package memory

import "testing"

func TestCartridgeCGBModeFlags(t *testing.T) {
	tests := []struct {
		name          string
		flag          byte
		wantSupported bool
		wantOnly      bool
		wantRunGBC    bool
	}{
		{name: "DMG only", flag: 0x00, wantSupported: false, wantOnly: false, wantRunGBC: false},
		{name: "CGB compatible defaults to DMG mode", flag: 0x80, wantSupported: true, wantOnly: false, wantRunGBC: false},
		{name: "CGB only forces GBC mode", flag: 0xC0, wantSupported: true, wantOnly: true, wantRunGBC: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cart := &CartContext{cgbFlag: tt.flag}
			if got := cart.IsGBCCart(); got != tt.wantSupported {
				t.Fatalf("IsGBCCart() = %v, want %v", got, tt.wantSupported)
			}
			if got := cart.IsGBCOnly(); got != tt.wantOnly {
				t.Fatalf("IsGBCOnly() = %v, want %v", got, tt.wantOnly)
			}
			if got := cart.RunsInGBCMode(); got != tt.wantRunGBC {
				t.Fatalf("RunsInGBCMode() = %v, want %v", got, tt.wantRunGBC)
			}
		})
	}
}
