package ui

// Existing methods...

// IncrementLY increments the LY register and checks for LY compare interrupt
func (p *PpuContext) IncrementLY() {
	lcdCtx := LcdCtx()
	lcdCtx.Ly++

	if lcdCtx.Ly == lcdCtx.LyCompare {
		LCDSLycSet(true)

		if LCDSStatInt(SSLyc) {
			// Request LCD STAT interrupt
			//p.externalPins.RequestInterrupt(cpu.IT_LCD_STAT)
		}
	} else {
		LCDSLycSet(false)
	}
}

// LoadLineSprites loads the sprites for the current line
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

// Other methods like PPUModeOam, PPUModeXfer, etc., should also use locking where they modify shared state.
