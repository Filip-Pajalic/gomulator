package cpu

import (
	"fmt"
)

// Assume necessary imports for bus, lcd packages

func WindowVisible() bool {
	//Always true, soemthing fishLy
	return LCDCWinEnable() && LcdCtx.WinX >= 0 &&
		LcdCtx.WinX <= 166 && LcdCtx.WinY >= 0 &&
		LcdCtx.WinY < YRES
}

func PixelFifoPush(value uint32) {
	Next := &FifoEntry{
		Next:  nil,
		Value: value,
	}

	if PpuCtx.Pfc.PixelFifo.head == nil {
		PpuCtx.Pfc.PixelFifo.head = Next
		PpuCtx.Pfc.PixelFifo.tail = Next
	} else {
		PpuCtx.Pfc.PixelFifo.tail.Next = Next
		PpuCtx.Pfc.PixelFifo.tail = Next
	}

	PpuCtx.Pfc.PixelFifo.size++
}

func PixelFifoPop() uint32 {
	if PpuCtx.Pfc.PixelFifo.size <= 0 {
		fmt.Println("ERR IN PIXEL FIFO!")
		panic(-8)
	}

	popped := PpuCtx.Pfc.PixelFifo.head
	PpuCtx.Pfc.PixelFifo.head = popped.Next
	PpuCtx.Pfc.PixelFifo.size--

	val := popped.Value
	// Explicit free is not needed in Go; memory is managed by the runtime.
	return val
}

func FetchSpritePixels(bit int, color uint32, bgColor uint8) uint32 {
	for i := 0; i < int(PpuCtx.FetchedEntryCount); i++ {
		spX := (PpuCtx.FetchedEntries[i].X - 8) +
			(LcdCtx.ScrollX % 8)

		if spX+8 < PpuCtx.Pfc.FifoX {
			continue
		}

		offset := PpuCtx.Pfc.FifoX - spX

		if offset < 0 || offset > 7 {
			continue
		}

		bit = int((7 - offset))
		//Double check
		if PpuCtx.FetchedEntries[i].FXFlip > 0 {
			bit = int(offset)
		}

		hi := (PpuCtx.Pfc.FetchEntryData[i*2] & (1 << bit)) >> bit
		lo := (PpuCtx.Pfc.FetchEntryData[(i*2)+1] & (1 << bit)) << 1

		bgPriority := PpuCtx.FetchedEntries[i].FBgp

		if hi|lo == 0 {
			continue
		}
		//is this wrong
		if bgPriority != 0 || bgColor == 0 {
			if PpuCtx.FetchedEntries[i].FPn > 0 {
				color = LcdCtx.Sp2Colors[hi|lo]
			} else {
				color = LcdCtx.Sp1Colors[hi|lo]
			}

			if hi|lo != 0 {
				break
			}
		}
	}

	return color
}

func PipelineFifoAdd() bool {
	if PpuCtx.Pfc.PixelFifo.size > 8 {
		return false
	}

	x := PpuCtx.Pfc.FetchX - (8 - (LcdCtx.ScrollX % 8))

	for i := 0; i < 8; i++ {
		bit := 7 - i
		hi := (PpuCtx.Pfc.BgwFetchData[1] & (1 << bit)) >> bit
		lo := (PpuCtx.Pfc.BgwFetchData[2] & (1 << bit)) << 1
		color := LcdCtx.BgColors[hi|lo]

		if !LCDCBGWEnable() {
			color = LcdCtx.BgColors[0]
		}

		if LCDCObjEnable() {
			color = FetchSpritePixels(bit, color, hi|lo)
		}

		if x >= 0 {
			PixelFifoPush(color)
			PpuCtx.Pfc.FifoX++
		}
	}

	return true
}

func PipelineLoadSpriteTile() {
	le := PpuCtx.LineSprites

	for le != nil {
		spX := (le.Entry.X - 8) + (LcdCtx.ScrollX % 8)

		if (spX >= PpuCtx.Pfc.FetchX && spX < PpuCtx.Pfc.FetchX+8) ||
			((spX+8) >= PpuCtx.Pfc.FetchX && (spX+8) < PpuCtx.Pfc.FetchX+8) {
			PpuCtx.FetchedEntries[PpuCtx.FetchedEntryCount] = le.Entry
			PpuCtx.FetchedEntryCount++
		}

		le = le.Next

		if le == nil || PpuCtx.FetchedEntryCount >= 3 {
			break
		}
	}
}

func PipelineLoadSpriteData(offset uint8) {
	curY := LcdCtx.Ly
	spriteHeight := LCDCObjHeight()

	for i := 0; i < int(PpuCtx.FetchedEntryCount); i++ {
		ty := ((curY + 16) - PpuCtx.FetchedEntries[i].Y) * 2

		if PpuCtx.FetchedEntries[i].FYFlip > 0 {
			ty = ((spriteHeight * 2) - 2) - ty
		}

		tileIndex := PpuCtx.FetchedEntries[i].Tile

		if spriteHeight == 16 {
			tileIndex &= ^uint8(1)
		}

		PpuCtx.Pfc.FetchEntryData[byte((i*2))+offset] =
			BusRead(0x8000 + (uint16(tileIndex) * 16) + uint16(ty) + uint16(offset))
	}
}

func PipelineLoadWindowTile() {
	if !WindowVisible() {
		return
	}

	windowY := LcdCtx.WinY

	if PpuCtx.Pfc.FetchX+7 >= LcdCtx.WinX &&
		PpuCtx.Pfc.FetchX+7 < LcdCtx.WinX+YRES+14 {
		if LcdCtx.Ly >= windowY && LcdCtx.Ly < windowY+XRES {
			wTileY := PpuCtx.WindowLine / 8

			var addr uint16
			if LCDCWinMapArea() == 0x9800 {
				addr = LCDCWinMapArea() + uint16((PpuCtx.Pfc.FetchX+7-LcdCtx.WinX)/8) + uint16(wTileY*32)
			} else {
				addr = LCDCWinMapArea() + uint16((PpuCtx.Pfc.FetchX+7-LcdCtx.WinX)/8) + uint16(wTileY*32)
			}

			PpuCtx.Pfc.BgwFetchData[0] = BusRead(addr)

			/*		PpuCtx.Pfc.BgwFetchData[0] = BusRead(LCDCWinMapArea() +
					((PpuCtx.Pfc.FetchX + 7 - LcdCtx.WinX) / 8) +
					(uint16(wTileY) * 32))*/

			if LCDCBgMapArea() == 0x8800 {
				PpuCtx.Pfc.BgwFetchData[0] += 128
			}
		}
	}
}

func PipelineFetch() {
	switch PpuCtx.Pfc.CurFetchState {
	case FS_TILE:
		PpuCtx.FetchedEntryCount = 0

		// Fetch background and window tile data
		PpuCtx.Pfc.BgwFetchData[0] = 0 // Fetch the tile data from memory

		// Fetch sprite data if enabled
		PipelineLoadSpriteTile()
		PpuCtx.Pfc.CurFetchState = FS_DATA0
		PpuCtx.Pfc.FetchX += 8

	case FS_DATA0:
		PpuCtx.Pfc.BgwFetchData[1] = 0 // Fetch the tile data from memory
		PipelineLoadSpriteData(0)
		PpuCtx.Pfc.CurFetchState = FS_DATA1

	case FS_DATA1:
		PpuCtx.Pfc.BgwFetchData[2] = 0 // Fetch the tile data from memory
		PipelineLoadSpriteData(1)
		PpuCtx.Pfc.CurFetchState = FS_IDLE

	case FS_IDLE:
		PpuCtx.Pfc.CurFetchState = FS_PUSH

	case FS_PUSH:
		if PipelineFifoAdd() {
			PpuCtx.Pfc.CurFetchState = FS_TILE
		}
	}
}

// PipelinePushPixel pushes a pixel onto the video buffer.
func PipelinePushPixel() {
	if PpuCtx.Pfc.PixelFifo.size > 8 {
		pixelData := PixelFifoPop()

		if PpuCtx.Pfc.LineX >= (PpuCtx.Pfc.LineX % 8) {
			PpuCtx.VideoBuffer[PpuCtx.Pfc.PushedX+PpuCtx.Pfc.LineX*XRES] = pixelData
			PpuCtx.Pfc.PushedX++
		}

		PpuCtx.Pfc.LineX++
	}
}

// PipelineProcess processes the pipeline fetch and push operations.
func PipelineProcess() {
	PpuCtx.Pfc.MapY = (PpuCtx.Pfc.LineX + PpuCtx.Pfc.LineX)
	PpuCtx.Pfc.MapX = (PpuCtx.Pfc.FetchX + PpuCtx.Pfc.FetchX)
	PpuCtx.Pfc.TileY = ((PpuCtx.Pfc.LineX + PpuCtx.Pfc.LineX) % 8) * 2

	if PpuCtx.LineTicks&1 == 0 {
		PipelineFetch()
	}

	PipelinePushPixel()
}

// PipelineFifoReset resets the FIFO queue.
func PipelineFifoReset() {
	for PpuCtx.Pfc.PixelFifo.size > 0 {
		PixelFifoPop()
	}

	PpuCtx.Pfc.PixelFifo.head = nil
	PpuCtx.Pfc.PixelFifo.tail = nil
	PpuCtx.Pfc.PixelFifo.size = 0
}
