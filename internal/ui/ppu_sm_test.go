package ui

import (
	"app/internal/cpu"
	"testing"
)

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

func TestLcdReadReturnsSTATWithUnusedBitSet(t *testing.T) {
	resetPPUTestState()

	lcd := LcdCtx()
	lcd.Lcds = uint8(ModeOam) | 0x04

	if got := LcdRead(0xFF41); got != 0x86 {
		t.Fatalf("STAT read = %02X, want 86", got)
	}
}

func TestLcdWriteSTATPreservesHardwareOwnedBits(t *testing.T) {
	resetPPUTestState()

	lcd := LcdCtx()
	lcd.Lcds = uint8(ModeOam) | 0x04

	LcdWrite(0xFF41, 0xFF)
	if lcd.Lcds != 0x7E {
		t.Fatalf("STAT after enabling interrupts = %02X, want 7E", lcd.Lcds)
	}

	LcdWrite(0xFF41, 0x00)
	if lcd.Lcds != 0x06 {
		t.Fatalf("STAT after clearing interrupts = %02X, want 06", lcd.Lcds)
	}
}

func TestLcdDisableResetsPPUStateAndStopsTicks(t *testing.T) {
	resetPPUTestState()

	ppu := PpuCtx()
	lcd := LcdCtx()
	lcd.Lcdc = LCDC_DISPLAY_ENABLE | LCDC_TILE_DATA | LCDC_BG_ENABLE
	lcd.Ly = 77
	lcd.Lcds = uint8(ModeXfer)
	ppu.LineTicks = 123
	ppu.WindowLine = 9

	LcdWrite(0xFF40, LCDC_TILE_DATA|LCDC_BG_ENABLE)

	if lcd.Ly != 0 {
		t.Fatalf("LY after LCD disable = %d, want 0", lcd.Ly)
	}
	if LCDSMode() != ModeHBlank {
		t.Fatalf("mode after LCD disable = %d, want HBlank", LCDSMode())
	}
	if ppu.LineTicks != 0 {
		t.Fatalf("LineTicks after LCD disable = %d, want 0", ppu.LineTicks)
	}
	if ppu.WindowLine != 0 {
		t.Fatalf("WindowLine after LCD disable = %d, want 0", ppu.WindowLine)
	}

	ppu.PpuTick()
	if lcd.Ly != 0 || ppu.LineTicks != 0 {
		t.Fatalf("PPU advanced while LCD disabled: LY=%d LineTicks=%d", lcd.Ly, ppu.LineTicks)
	}
}

func TestSetLCDModeRequestsSTATInterrupt(t *testing.T) {
	resetPPUTestState()
	ctx := cpu.NewCpuContext(nil)

	LcdCtx().Lcds = uint8(SSOam)

	SetLCDMode(ModeOam)

	if ctx.IntFlags&byte(cpu.IT_LCD_STAT) == 0 {
		t.Fatalf("expected LCD STAT interrupt, IF=%02X", ctx.IntFlags)
	}
}

func TestPipelinePushPixelDiscardsBackgroundFineScroll(t *testing.T) {
	resetPPUTestState()

	ppu := NewPpuContext()
	lcd := LcdCtx()
	lcd.ScrollX = 3

	for i := 0; i < 9; i++ {
		ppu.PixelFifoPushWithIndex(0x11223344, 1, false)
	}

	ppu.PipelinePushPixel()

	if ppu.Pfc.PushedX != 0 {
		t.Fatalf("PushedX = %d, want 0 while discarding background fine scroll", ppu.Pfc.PushedX)
	}
	if ppu.Pfc.LineX != 1 {
		t.Fatalf("LineX = %d, want 1 after one discarded background pixel", ppu.Pfc.LineX)
	}
	if ppu.VideoBuffer[0] != 0 {
		t.Fatalf("first pixel = %08X, want untouched while discarding background fine scroll", ppu.VideoBuffer[0])
	}
}

func TestPipelinePushPixelDoesNotDiscardWindowForBackgroundFineScroll(t *testing.T) {
	resetPPUTestState()

	ppu := NewPpuContext()
	lcd := LcdCtx()
	lcd.ScrollX = 3

	for i := 0; i < 9; i++ {
		ppu.PixelFifoPushWithIndex(0x55667788, 1, true)
	}

	ppu.PipelinePushPixel()

	if ppu.Pfc.PushedX != 1 {
		t.Fatalf("PushedX = %d, want 1 for window pixel despite background fine scroll", ppu.Pfc.PushedX)
	}
	if ppu.Pfc.LineX != 1 {
		t.Fatalf("LineX = %d, want 1 after one window pixel", ppu.Pfc.LineX)
	}
	if ppu.VideoBuffer[0] != 0x55667788 {
		t.Fatalf("first pixel = %08X, want window pixel unaffected by SCX", ppu.VideoBuffer[0])
	}
}
