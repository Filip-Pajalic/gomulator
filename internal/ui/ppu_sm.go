package ui

import (
	"app/internal/cpu"
	logger "app/internal/logger"
)

func (p *PpuContext) IncrementLY() {
	currentLy := LcdCtx().Ly
	windowEnabled := LCDCWinEnable()
	windowOnScreen := LcdCtx().WinX < 167

	switch {
	case !windowEnabled || !windowOnScreen:
		p.WindowLine = 0
	case currentLy < LcdCtx().WinY:
		p.WindowLine = 0
	case currentLy < YRES:
		if currentLy == LcdCtx().WinY && p.WindowLine != 0 {
			p.WindowLine = 0
		}
		p.WindowLine++
		if p.WindowLine <= 20 {
			logger.Debug("PPU: Window line advanced to %d after scanline %d", p.WindowLine, currentLy)
		}
	}

	lcdCtx := LcdCtx()
	lcdCtx.Ly++

	if lcdCtx.Ly <= 20 {
		logger.Debug("PPU: LY incremented to %d, LYC=%d, LCDC=0x%02X", lcdCtx.Ly, lcdCtx.LyCompare, lcdCtx.Lcdc)
	}

	if lcdCtx.Ly == lcdCtx.LyCompare {
		LCDSLycSet(true)

		logger.Debug("PPU: LY=LYC match! LY=%d, LYC=%d, STAT=0x%02X", lcdCtx.Ly, lcdCtx.LyCompare, lcdCtx.Lcds)

		if LCDSStatInt(SSLyc) {
			// Request LCD STAT interrupt (like reference)
			cpu.CpuRequestInterrupt(cpu.IT_LCD_STAT)
			logger.Debug("PPU: *** LY=LYC INTERRUPT REQUESTED *** (LY=%d)", lcdCtx.Ly)
		} else {
			logger.Debug("PPU: LY=LYC match but STAT interrupt not enabled")
		}
	} else {
		LCDSLycSet(false)
	}
}

func (p *PpuContext) LoadLineSprites() {
	curY := LcdCtx().Ly
	spriteHeight := LCDCObjHeight()

	// Reset LineEntryArray and LineSpriteCount
	for i := range p.LineEntryArray {
		p.LineEntryArray[i] = OamLineEntry{}
	}
	p.LineSpriteCount = 0
	p.LineSprites = nil

	for i := 0; i < 40; i++ {
		e := p.OamRam[i]

		if e.X == 0 {
			// x = 0 means not visible...
			continue
		}

		if p.LineSpriteCount >= 10 {
			// max 10 sprites per line...
			break
		}

		if e.Y <= curY+16 && e.Y+uint8(spriteHeight) > curY+16 {
			// this sprite is on the current line.
			entry := &p.LineEntryArray[p.LineSpriteCount]
			p.LineSpriteCount++

			entry.Entry = e
			entry.Next = nil

			if p.LineSprites == nil || p.LineSprites.Entry.X > e.X {
				entry.Next = p.LineSprites
				p.LineSprites = entry
				continue
			}

			// Insert sprite in sorted order based on X position
			le := p.LineSprites
			var prev *OamLineEntry

			for le != nil {
				if le.Entry.X > e.X {
					if prev != nil {
						prev.Next = entry
					} else {
						p.LineSprites = entry
					}
					entry.Next = le
					break
				}

				if le.Next == nil {
					le.Next = entry
					break
				}

				prev = le
				le = le.Next
			}
		}
	}
}
