package ui

import "testing"

func resetPPUTestState() {
	lcdContext = LcdContext{}
	ppuInstance = nil
}

func TestIncrementLYPreservesWindowLineWhenWXIsOffscreen(t *testing.T) {
	resetPPUTestState()

	ppu := NewPpuContext()
	ppu.WindowLine = 16
	lcd := LcdCtx()
	lcd.Lcdc = LCDC_DISPLAY_ENABLE | LCDC_WIN_ENABLE | LCDC_BG_ENABLE
	lcd.Ly = 40
	lcd.LyCompare = 0xFF
	lcd.WinY = 0
	lcd.WinX = 167

	ppu.IncrementLY()

	if ppu.WindowLine != 16 {
		t.Fatalf("WindowLine = %d, want 16 when WX is offscreen", ppu.WindowLine)
	}
	if lcd.Ly != 41 {
		t.Fatalf("LY = %d, want 41", lcd.Ly)
	}
}

func TestIncrementLYAdvancesWindowLineWhenWindowIsVisible(t *testing.T) {
	resetPPUTestState()

	ppu := NewPpuContext()
	ppu.WindowLine = 16
	lcd := LcdCtx()
	lcd.Lcdc = LCDC_DISPLAY_ENABLE | LCDC_WIN_ENABLE | LCDC_BG_ENABLE
	lcd.Ly = 40
	lcd.LyCompare = 0xFF
	lcd.WinY = 0
	lcd.WinX = 7

	ppu.IncrementLY()

	if ppu.WindowLine != 17 {
		t.Fatalf("WindowLine = %d, want 17 when window is visible", ppu.WindowLine)
	}
}

func TestIncrementLYResetsWindowLineBeforeWindowY(t *testing.T) {
	resetPPUTestState()

	ppu := NewPpuContext()
	ppu.WindowLine = 16
	lcd := LcdCtx()
	lcd.Lcdc = LCDC_DISPLAY_ENABLE | LCDC_WIN_ENABLE | LCDC_BG_ENABLE
	lcd.Ly = 4
	lcd.LyCompare = 0xFF
	lcd.WinY = 8
	lcd.WinX = 7

	ppu.IncrementLY()

	if ppu.WindowLine != 0 {
		t.Fatalf("WindowLine = %d, want reset before WY", ppu.WindowLine)
	}
}
