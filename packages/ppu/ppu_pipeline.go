// ppu_pipeline.go
package ppu

import (
	"pajalic.go.emulator/packages/logger"
	"pajalic.go.emulator/packages/pubsub"
	"pajalic.go.emulator/packages/ui"
)

// Other imports...

// WindowVisible checks if the window should be visible based on LCDC settings
func (p *PpuContext) WindowVisible() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return ui.LCDCWinEnable() &&
		ui.LcdCtx().WinX >= 0 &&
		ui.LcdCtx().WinX <= 166 &&
		ui.LcdCtx().WinY >= 0 &&
		ui.LcdCtx().WinY < YRES
}

// PixelFifoPush pushes a pixel value onto the FIFO queue
func (p *PpuContext) PixelFifoPush(value uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	next := &FifoEntry{
		Next:  nil,
		Value: value,
	}

	if p.Pfc.PixelFifo.head == nil {
		p.Pfc.PixelFifo.head = next
		p.Pfc.PixelFifo.tail = next
	} else {
		p.Pfc.PixelFifo.tail.Next = next
		p.Pfc.PixelFifo.tail = next
	}

	p.Pfc.PixelFifo.size++
}

// PixelFifoPop pops a pixel value from the FIFO queue
func (p *PpuContext) PixelFifoPop() uint32 {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Pfc.PixelFifo.size <= 0 {
		logger.Error("PPU PixelFifoPop: FIFO is empty!")
		return 0xFF
	}

	popped := p.Pfc.PixelFifo.head
	p.Pfc.PixelFifo.head = popped.Next
	p.Pfc.PixelFifo.size--

	val := popped.Value
	// Memory is managed by Go's runtime; no explicit free needed
	return val
}

// FetchSpritePixels fetches sprite pixels considering various flags and priorities
func (p *PpuContext) FetchSpritePixels(bit int, color uint32, bgColor uint8) uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for i := 0; i < int(p.FetchedEntryCount); i++ {
		spX := (p.FetchedEntries[i].X - 8) + (ui.LcdCtx().ScrollX % 8)

		if spX+8 < p.Pfc.FifoX {
			continue
		}

		offset := p.Pfc.FifoX - spX

		if offset < 0 || offset > 7 {
			continue
		}

		bit = int((7 - offset))
		if p.FetchedEntries[i].FXFlip > 0 {
			bit = int(offset)
		}

		hi := (p.Pfc.FetchEntryData[i*2] & (1 << bit)) >> bit
		lo := (p.Pfc.FetchEntryData[(i*2)+1] & (1 << bit)) << 1

		bgPriority := p.FetchedEntries[i].FBgp

		if hi|lo == 0 {
			continue
		}

		if bgPriority != 0 || bgColor == 0 {
			if p.FetchedEntries[i].FPn > 0 {
				color = ui.LcdCtx().Sp2Colors[hi|lo]
			} else {
				color = ui.LcdCtx().Sp1Colors[hi|lo]
			}

			if hi|lo != 0 {
				break
			}
		}
	}

	return color
}

// PipelineFifoAdd adds pixels to the FIFO pipeline if there's space
func (p *PpuContext) PipelineFifoAdd() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Pfc.PixelFifo.size > 8 {
		return false
	}

	x := p.Pfc.FetchX - (8 - (ui.LcdCtx().ScrollX % 8))

	for i := 0; i < 8; i++ {
		bit := 7 - i
		hi := (p.Pfc.BgwFetchData[1] & (1 << bit)) >> bit
		lo := (p.Pfc.BgwFetchData[2] & (1 << bit)) << 1
		color := ui.LcdCtx().BgColors[hi|lo]

		if !ui.LCDCObjEnable() {
			color = ui.LcdCtx().BgColors[0]
		}

		if ui.LCDCObjEnable() {
			color = p.FetchSpritePixels(bit, color, hi|lo)
		}

		if x >= 0 {
			p.PixelFifoPush(color)
			p.Pfc.FifoX++
		}
	}

	return true
}

// PipelineLoadSpriteTile loads sprite tile data into fetched entries
func (p *PpuContext) PipelineLoadSpriteTile() {
	p.mu.Lock()
	defer p.mu.Unlock()

	le := p.LineSprites

	for le != nil {
		spX := (le.Entry.X - 8) + (ui.LcdCtx().ScrollX % 8)

		if (spX >= p.Pfc.FetchX && spX < p.Pfc.FetchX+8) ||
			((spX+8) >= p.Pfc.FetchX && (spX+8) < p.Pfc.FetchX+8) {
			p.FetchedEntries[p.FetchedEntryCount] = le.Entry
			p.FetchedEntryCount++
		}

		le = le.Next

		if le == nil || p.FetchedEntryCount >= 3 {
			break
		}
	}
}

// PipelineLoadSpriteData loads sprite data based on the current fetch offset
func (p *PpuContext) PipelineLoadSpriteData(offset uint8) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	curY := ui.LcdCtx().Ly
	spriteHeight := ui.LCDCObjHeight()

	for i := 0; i < int(p.FetchedEntryCount); i++ {
		ty := ((curY + 16) - p.FetchedEntries[i].Y) * 2

		if p.FetchedEntries[i].FYFlip > 0 {
			ty = ((spriteHeight * 2) - 2) - ty
		}

		tileIndex := p.FetchedEntries[i].Tile

		if spriteHeight == 16 {
			tileIndex &= ^uint8(1)
		}

		p.Pfc.FetchEntryData[byte((i*2))+offset] =
			pubsub.BusCtx().BusRead(0x8000 + (uint16(tileIndex) * 16) + uint16(ty) + uint16(offset))
	}
}

// PipelineLoadWindowTile loads window tile data if the window is visible
func (p *PpuContext) PipelineLoadWindowTile() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.WindowVisible() {
		return
	}

	windowY := ui.LcdCtx().WinY

	if p.Pfc.FetchX+7 >= ui.LcdCtx().WinX &&
		p.Pfc.FetchX+7 < ui.LcdCtx().WinX+YRES+14 {
		if ui.LcdCtx().Ly >= windowY && ui.LcdCtx().Ly < windowY+XRES {
			wTileY := p.WindowLine / 8

			var addr uint16
			if ui.LCDCWinMapArea() == 0x9800 {
				addr = ui.LCDCWinMapArea() + uint16((p.Pfc.FetchX+7-ui.LcdCtx().WinX)/8) + uint16(wTileY*32)
			} else {
				addr = ui.LCDCWinMapArea() + uint16((p.Pfc.FetchX+7-ui.LcdCtx().WinX)/8) + uint16(wTileY*32)
			}

			p.Pfc.BgwFetchData[0] = pubsub.BusCtx().BusRead(addr)

			if ui.LCDCWinMapArea() == 0x8800 {
				p.Pfc.BgwFetchData[0] += 128
			}
		}
	}
}

// PipelineFetch handles the fetch phase of the PPU pipeline
func (p *PpuContext) PipelineFetch() {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.Pfc.CurFetchState {
	case FS_TILE:
		p.FetchedEntryCount = 0

		// Fetch background and window tile data
		p.Pfc.BgwFetchData[0] = 0 // Initialize fetch data

		// Fetch sprite data if enabled
		p.PipelineLoadSpriteTile()
		p.Pfc.CurFetchState = FS_DATA0
		p.Pfc.FetchX += 8

	case FS_DATA0:
		p.Pfc.BgwFetchData[1] = 0 // Initialize fetch data
		p.PipelineLoadSpriteData(0)
		p.Pfc.CurFetchState = FS_DATA1

	case FS_DATA1:
		p.Pfc.BgwFetchData[2] = 0 // Initialize fetch data
		p.PipelineLoadSpriteData(1)
		p.Pfc.CurFetchState = FS_IDLE

	case FS_IDLE:
		p.Pfc.CurFetchState = FS_PUSH

	case FS_PUSH:
		if p.PipelineFifoAdd() {
			p.Pfc.CurFetchState = FS_TILE
		}
	}
}

// PipelinePushPixel pushes a pixel onto the video buffer
func (p *PpuContext) PipelinePushPixel() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Pfc.PixelFifo.size > 8 {
		pixelData := p.PixelFifoPop()

		if p.Pfc.LineX >= (p.Pfc.LineX % 8) {
			p.VideoBuffer[p.Pfc.PushedX+p.Pfc.LineX*XRES] = pixelData
			p.Pfc.PushedX++
		}

		p.Pfc.LineX++
	}
}

// PipelineProcess processes the pipeline fetch and push operations
func (p *PpuContext) PipelineProcess() {
	p.mu.Lock()
	p.Pfc.MapY = (p.Pfc.LineX + p.Pfc.LineX)
	p.Pfc.MapX = (p.Pfc.FetchX + p.Pfc.FetchX)
	p.Pfc.TileY = ((p.Pfc.LineX + p.Pfc.LineX) % 8) * 2
	p.mu.Unlock()

	if p.LineTicks&1 == 0 {
		p.PipelineFetch()
	}

	p.PipelinePushPixel()
}

// PipelineFifoReset resets the FIFO queue
func (p *PpuContext) PipelineFifoReset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for p.Pfc.PixelFifo.size > 0 {
		p.PixelFifoPop()
	}

	p.Pfc.PixelFifo.head = nil
	p.Pfc.PixelFifo.tail = nil
	p.Pfc.PixelFifo.size = 0
}
