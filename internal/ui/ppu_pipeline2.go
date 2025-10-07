package ui

import (
	"app/internal/logger"
)

func (p *PpuContext) PipelineTick() {
	p.Pfc.MapY += 1
	p.Pfc.MapX += 1

	// Process fetcher
	p.PixelFetch()

	// Push pixels to screen
	p.PipelinePushPixel()
}

func (p *PpuContext) PipelineReset() {
	p.Pfc.CurFetchState = FS_TILE
	p.Pfc.LineX = 0
	p.Pfc.PushedX = 0
	p.Pfc.FetchX = 0
	p.Pfc.FifoX = 0

	// Clear FIFO
	p.Pfc.PixelFifo.size = 0
	p.Pfc.PixelFifo.head = 0
	p.Pfc.PixelFifo.tail = 0

	logger.Debug("Pipeline reset for line %d", LcdCtx().Ly)
}

func (p *PpuContext) PipelineLoadWindowTile() {
	// Window tile loading - similar to background but using window coordinates
	if !LCDCWinEnable() {
		logger.Debug("Window disabled, skipping window tile load")
		return
	}

	// Check if window should be visible at current position
	winX := LcdCtx().WinX
	winY := LcdCtx().WinY
	currentX := p.Pfc.FetchX
	currentY := LcdCtx().Ly

	// Window is visible if current position is within window bounds
	if currentY >= winY && currentX >= winX-7 {
		// Calculate window-relative coordinates
		windowTileY := int(p.WindowLine) / 8
		windowTileX := int(currentX-(winX-7)) / 8

		// Get window map address
		mapAddr := LCDCWinMapArea() + uint16(windowTileY*32+windowTileX)

		// Read tile number from window map
		tileNum := p.VramRead(mapAddr)

		// Handle signed tile numbers for window (same as background)
		if LCDCBGWDataArea() == 0x8800 {
			tileNum += 128
		}

		// Store in fetch data
		p.Pfc.BgwFetchData[0] = tileNum

		logger.Debug("PPU: Window tile loaded - tileNum=0x%02X at (%d,%d)", tileNum, windowTileX, windowTileY)
	}
}

func (p *PpuContext) PipelineLoadSpriteTile() {
	// TODO: Implement sprite tile loading
	logger.Debug("Sprite tile loading not yet implemented")
}

func (p *PpuContext) PipelineLoadSpriteData(offset int) {
	// TODO: Implement sprite data loading
	logger.Debug("Sprite data loading not yet implemented for offset %d", offset)
}
