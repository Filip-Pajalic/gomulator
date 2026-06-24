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

func TestMBC1ROMBankSelectionWrapsToAvailableBanks(t *testing.T) {
	const bankSize = 0x4000

	rom := make([]byte, 4*bankSize)
	for bank := 0; bank < 4; bank++ {
		rom[bank*bankSize] = byte(0xA0 + bank)
	}

	cart := &CartContext{romData: rom, romBank: 1}

	cart.CartWrite(0x2000, 0x00)
	if got := cart.CartRead(0x4000); got != 0xA1 {
		t.Fatalf("bank 0 read = %02X, want A1", got)
	}

	cart.CartWrite(0x2000, 0x02)
	if got := cart.CartRead(0x4000); got != 0xA2 {
		t.Fatalf("bank 2 read = %02X, want A2", got)
	}

	cart.CartWrite(0x2000, 0x04)
	if got := cart.CartRead(0x4000); got != 0xA0 {
		t.Fatalf("masked bank 4 read = %02X, want A0", got)
	}

	cart.CartWrite(0x2000, 0x1F)
	if got := cart.CartRead(0x4000); got != 0xA3 {
		t.Fatalf("masked bank 31 read = %02X, want A3", got)
	}

	cart.CartWrite(0x4000, 0x03)
	if got := cart.CartRead(0x4000); got != 0xA3 {
		t.Fatalf("secondary bank register should be ignored for 4-bank ROM, got %02X, want A3", got)
	}
}
