package ppu

import (
	"fmt"
	"pajalic.go.emulator/packages/cpu"
	"pajalic.go.emulator/packages/ui"
	"time"
)

// IncrementLY increments the LY register and checks for LY compare interrupt
func IncrementLY() {
	lcdCtx := ui.LcdGetContext()
	lcdCtx.Ly++

	if lcdCtx.Ly == lcdCtx.LyCompare {
		ui.LCDSLycSet(true)

		if ui.LCDSStatInt(ui.SSLyc) {
			cpu.CpuRequestInterrupt(cpu.IT_LCD_STAT)
		}
	} else {
		ui.LCDSLycSet(false)
	}
}

// LoadLineSprites loads the sprites for the current line
func LoadLineSprites() {

	curY := ui.LcdCtx.Ly
	spriteHeight := ui.LCDCObjHeight()
	//This is probably bad
	PpuCtx.LineEntryArray = [10]OamLineEntry(make([]OamLineEntry, 40))
	PpuCtx.LineSpriteCount = 0

	for i := 0; i < 40; i++ {
		e := PpuCtx.OamRam[i]

		if e.X == 0 {
			// x = 0 means not visible...
			continue
		}

		if PpuCtx.LineSpriteCount >= 10 {
			// max 10 sprites per line...
			break
		}

		if e.Y <= curY+16 && e.Y+spriteHeight > curY+16 {
			// this sprite is on the current line.
			entry := &PpuCtx.LineEntryArray[PpuCtx.LineSpriteCount]
			PpuCtx.LineSpriteCount++

			entry.Entry = e
			entry.Next = nil

			if PpuCtx.LineSprites == nil || PpuCtx.LineSprites.Entry.X > e.X {
				entry.Next = PpuCtx.LineSprites
				PpuCtx.LineSprites = entry
				continue
			}

			// do some sorting...
			le := PpuCtx.LineSprites
			var prev *OamLineEntry

			for le != nil {
				if le.Entry.X > e.X {
					prev.Next = entry
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

// PPUModeOam handles the PPU OAM mode
func PPUModeOam() {
	ppuCtx := PpuCtx

	if ppuCtx.LineTicks >= 80 {
		ui.LCDSModeSet(ui.ModeXfer)
		ppuCtx.Pfc.CurFetchState = FS_TILE
		ppuCtx.Pfc.LineX = 0
		ppuCtx.Pfc.FetchX = 0
		ppuCtx.Pfc.PushedX = 0
		ppuCtx.Pfc.FifoX = 0
	}

	if ppuCtx.LineTicks == 1 {
		// read oam on the first tick only...
		ppuCtx.LineSprites = nil
		ppuCtx.LineSpriteCount = 0
		LoadLineSprites()
	}
}

// PPUModeXfer handles the PPU transfer mode
func PPUModeXfer() {
	PipelineProcess()

	if PpuCtx.Pfc.PushedX >= XRES {
		PipelineFifoReset()
		ui.LCDSModeSet(ui.ModeHBlank)

		if ui.LCDSStatInt(ui.SSHBlank) {
			cpu.CpuRequestInterrupt(cpu.IT_LCD_STAT)
		}
	}
}

// PPUModeVblank handles the PPU VBlank mode
func PPUModeVblank() {

	if PpuCtx.LineTicks >= TICKS_PER_LINE {
		IncrementLY()

		if ui.LcdCtx.Ly >= LINES_PER_FRAME {
			ui.LCDSModeSet(ui.ModeOam)
			ui.LcdCtx.Ly = 0
		}

		PpuCtx.LineTicks = 0
	}
}

var (
	targetFrameTime = 1000 / 60
	prevFrameTime   = time.Now().UnixMilli()
	startTimer      = time.Now().UnixMilli()
	frameCount      = 0
)

// PPUModeHblank handles the PPU HBlank mode
func PPUModeHblank() {

	if PpuCtx.LineTicks >= TICKS_PER_LINE {
		IncrementLY()

		if ui.LcdCtx.Ly >= YRES {
			ui.LCDSModeSet(ui.ModeVBlank)

			cpu.CpuRequestInterrupt(cpu.IT_VBLANK)

			if ui.LCDSStatInt(ui.SSVBlank) {
				cpu.CpuRequestInterrupt(cpu.IT_LCD_STAT)
			}

			PpuCtx.CurrentFrame++

			// calculate FPS...
			end := time.Now().UnixMilli()
			frameTime := end - prevFrameTime

			if frameTime < int64(targetFrameTime) {
				time.Sleep(time.Duration(int64(targetFrameTime)-frameTime) * time.Millisecond)
			}

			if end-startTimer >= 1000 {
				fps := frameCount
				startTimer = end
				frameCount = 0

				fmt.Printf("FPS: %d\n", fps)
			}

			frameCount++
			prevFrameTime = time.Now().UnixMilli()
		} else {
			ui.LCDSModeSet(ui.ModeOam)
		}

		PpuCtx.LineTicks = 0
	}
}
