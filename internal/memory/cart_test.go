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

	cart := &CartContext{romData: rom, header: &romHeader{CartType: 0x01}, romBank: 1}

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

func TestMBC2ROMBankSelectionUsesAddressBit8(t *testing.T) {
	cart := newMBC2TestCart(0x05, 8)

	cart.CartWrite(0x0000, 0x0A)
	if !cart.ramEnabled {
		t.Fatalf("RAM enable write with bit 8 clear did not enable RAM")
	}

	cart.CartWrite(0x0100, 0x03)
	if got := cart.CurrentROMBank(); got != 3 {
		t.Fatalf("ROM bank after bit 8 set write = %d, want 3", got)
	}
	if got := cart.CartRead(0x4000); got != 0xB3 {
		t.Fatalf("bank 3 read = %02X, want B3", got)
	}

	cart.CartWrite(0x0000, 0x02)
	if got := cart.CurrentROMBank(); got != 3 {
		t.Fatalf("RAM control write changed ROM bank to %d, want 3", got)
	}
	if cart.ramEnabled {
		t.Fatalf("RAM disable write with bit 8 clear left RAM enabled")
	}

	cart.CartWrite(0x2100, 0x00)
	if got := cart.CurrentROMBank(); got != 1 {
		t.Fatalf("bank 0 remap = %d, want 1", got)
	}

	cart.CartWrite(0x2100, 0x0F)
	if got := cart.CurrentROMBank(); got != 7 {
		t.Fatalf("bank wrap = %d, want 7", got)
	}
}

func TestMBC2RAMEnableAndNibbleRAM(t *testing.T) {
	cart := newMBC2TestCart(0x05, 2)

	cart.CartWrite(0xA000, 0x0F)
	if got := cart.CartRead(0xA000); got != 0xFF {
		t.Fatalf("disabled RAM read = %02X, want FF", got)
	}

	cart.CartWrite(0x0000, 0x0A)
	cart.CartWrite(0xA000, 0xAB)
	if got := cart.CartRead(0xA000); got != 0xFB {
		t.Fatalf("MBC2 low nibble read = %02X, want FB", got)
	}
	if got := cart.CartRead(0xA200); got != 0xFB {
		t.Fatalf("MBC2 RAM echo read = %02X, want FB", got)
	}

	cart.CartWrite(0x0000, 0x00)
	cart.CartWrite(0xA000, 0x05)
	cart.CartWrite(0x0000, 0x0A)
	if got := cart.CartRead(0xA000); got != 0xFB {
		t.Fatalf("disabled write changed RAM to %02X, want FB", got)
	}
}

func TestMBC2InitializesInternalRAMWithoutHeaderRAM(t *testing.T) {
	cart := &CartContext{header: &romHeader{CartType: 0x05, RamSize: 0x00}}
	cart.initializeRAM()

	if len(cart.mbc2RAM) != 512 {
		t.Fatalf("MBC2 RAM length = %d, want 512", len(cart.mbc2RAM))
	}
	if len(cart.ramData) != 0 {
		t.Fatalf("MBC2 external RAM length = %d, want 0", len(cart.ramData))
	}
}

func TestMBC2BatterySaveLoadAndFlush(t *testing.T) {
	store := &fakeSaveStore{data: map[string][]byte{
		"save-key": {0x01, 0x02, 0x03},
	}}
	restore := SetSaveStore(store)
	defer SetSaveStore(restore)

	cart := newMBC2TestCart(0x06, 2)
	cart.saveKey = "save-key"
	cart.loadBatteryBackedRAM()

	if got := cart.mbc2RAM[0]; got != 0x01 {
		t.Fatalf("loaded RAM[0] = %02X, want 01", got)
	}
	if got := cart.mbc2RAM[2]; got != 0x03 {
		t.Fatalf("loaded RAM[2] = %02X, want 03", got)
	}
	if got := cart.mbc2RAM[3]; got != 0x00 {
		t.Fatalf("short save should zero-fill RAM[3], got %02X", got)
	}

	cart.CartWrite(0x0000, 0x0A)
	cart.CartWrite(0xA000, 0x0C)
	if !cart.saveDirty {
		t.Fatalf("MBC2+BATTERY RAM write did not mark save dirty")
	}

	cart.FlushSave()
	saved := store.data["save-key"]
	if len(saved) != 512 {
		t.Fatalf("saved length = %d, want 512", len(saved))
	}
	if got := saved[0]; got != 0x0C {
		t.Fatalf("saved RAM[0] = %02X, want 0C", got)
	}
	if cart.saveDirty {
		t.Fatalf("FlushSave left saveDirty set")
	}
}

func TestMBC2WithoutBatteryDoesNotPersist(t *testing.T) {
	store := &fakeSaveStore{data: map[string][]byte{}}
	restore := SetSaveStore(store)
	defer SetSaveStore(restore)

	cart := newMBC2TestCart(0x05, 2)
	cart.saveKey = "volatile"
	cart.CartWrite(0x0000, 0x0A)
	cart.CartWrite(0xA000, 0x07)

	if cart.saveDirty {
		t.Fatalf("MBC2 without battery marked save dirty")
	}
	cart.FlushSave()
	if _, ok := store.data["volatile"]; ok {
		t.Fatalf("MBC2 without battery wrote a save")
	}
}

func newMBC2TestCart(cartType byte, banks int) *CartContext {
	const bankSize = 0x4000

	rom := make([]byte, banks*bankSize)
	for bank := 0; bank < banks; bank++ {
		rom[bank*bankSize] = byte(0xB0 + bank)
	}

	cart := &CartContext{
		romData: rom,
		header:  &romHeader{CartType: cartType},
		romBank: 1,
	}
	cart.initializeRAM()
	return cart
}

type fakeSaveStore struct {
	data map[string][]byte
}

func (s *fakeSaveStore) Load(key string) ([]byte, error) {
	data := s.data[key]
	return append([]byte(nil), data...), nil
}

func (s *fakeSaveStore) Save(key string, data []byte) error {
	s.data[key] = append([]byte(nil), data...)
	return nil
}
