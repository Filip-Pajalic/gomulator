package ui

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	emu "pajalic.go.emulator/packages/cpu"
	log "pajalic.go.emulator/packages/logger"
	"unsafe"
)

var SCREEN_WIDTH int32 = 1024

var SCREEN_HEIGHT int32 = 768

var sdlWindow *sdl.Window

var sdlRenderer *sdl.Renderer

var sdlTexture *sdl.Texture

var screen *sdl.Surface

var sdlDebugWindow *sdl.Window

var sdlDebugRenderer *sdl.Renderer

var sdlDebugTexture *sdl.Texture

var debugScreen *sdl.Surface

var scale int32 = 4

func UiInit() {

	log.Info("Cart loaded..")
	sdl.Init(sdl.INIT_VIDEO)
	log.Info("SDL INIT")
	ttf.Init()
	log.Info("TTF INIT")
	var err error
	sdlWindow, sdlRenderer, err = sdl.CreateWindowAndRenderer(SCREEN_WIDTH, SCREEN_HEIGHT, sdl.WINDOW_RESIZABLE)
	if err != nil {
		log.Fatal(err.Error())
	}

	sdlDebugWindow, sdlDebugRenderer, err = sdl.CreateWindowAndRenderer(16*8*scale, 32*8*scale, 0)
	if err != nil {
		log.Fatal(err.Error())
	}

	debugScreen, _ = sdl.CreateRGBSurface(0, (16*8*scale)+(16*scale),
		(32*8*scale)+(64*scale), 32,
		0x00FF0000,
		0x0000FF00,
		0x000000FF,
		0xFF000000)

	sdlDebugTexture, _ = sdlDebugRenderer.CreateTexture(
		sdl.PIXELFORMAT_ABGR8888,
		sdl.TEXTUREACCESS_STREAMING,
		(16*8*scale)+(16*scale),
		(32*8*scale)+(64*scale))
	if err != nil {
		log.Fatal(err.Error())
	}

	x, y := sdlWindow.GetPosition()
	sdlDebugWindow.SetPosition(x+SCREEN_WIDTH+10, y)

}

func DestroyWindow() {
	sdlWindow.Destroy()
	sdlDebugWindow.Destroy()
}

func delay(ms uint32) {
	sdl.Delay(ms)
}

// Might be wrong
var tileColors = [4]uint32{0xFFFFFFFF, 0xFFAAAAAA, 0xFF555555, 0xFF000000}

func displayTile(surface *sdl.Surface, startLocation uint16, tileNum uint16, x int32, y int32) {
	var rc = sdl.Rect{}

	for tileY := int32(0); tileY < 16; tileY += 2 {
		var b1 = emu.BusRead(startLocation + (tileNum * 16) + uint16(tileY))
		var b2 = emu.BusRead(startLocation + (tileNum * 16) + uint16(tileY) + 1)

		for bit := int32(7); bit >= 0; bit-- {
			var b1bit byte = 0
			var b2bit byte = 0
			if b1&(1<<bit) == 1 {
				b1bit = 1
			}
			if b2&(1<<bit) == 1 {
				b2bit = 1
			}

			var hi = b1bit << 1
			var lo = b2bit

			var color = hi | lo

			rc.X = x + ((7 - bit) * scale)
			rc.Y = y + (tileY / 2 * scale)
			rc.W = scale
			rc.H = scale
			surface.FillRect(&rc, tileColors[color])
		}
	}
}

func UpdateDbgWindows() {
	var xDraw int32 = 0
	var yDraw int32 = 0
	var tileNum uint16 = 0

	var rc = sdl.Rect{}
	rc.X = 0
	rc.Y = 0
	rc.W = debugScreen.W
	rc.H = debugScreen.H
	debugScreen.FillRect(&rc, 0xFF111111)

	var addr uint16 = 0x8000

	//384 tiles, 24 x 16
	for y := int32(0); y < 24; y++ {
		for x := int32(0); x < 16; x++ {
			displayTile(debugScreen, addr, tileNum, xDraw+(x*scale), yDraw+(y*scale))
			xDraw += 8 * scale
			tileNum++
		}

		yDraw += 8 * scale
		xDraw = 0
	}
	pixels := debugScreen.Pixels()

	sdlDebugTexture.Update(nil, unsafe.Pointer(&pixels[0]), int(debugScreen.Pitch))
	sdlDebugRenderer.Clear()
	sdlDebugRenderer.Copy(sdlDebugTexture, nil, nil)

	sdlDebugRenderer.Present()
}

func UiUpdate() {
	UpdateDbgWindows()
}

func UiHandleEvents() {
	//Quit both windows
	for e := sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
		//TODO SDL_UpdateWindowSurface(sdlWindow);
		//TODO SDL_UpdateWindowSurface(sdlTraceWindow);
		//TODO SDL_UpdateWindowSurface(sdlDebugWindow);

		switch e.(type) {
		case *sdl.QuitEvent:
			println("Quit")
			emu.GetEmuContext().Die = true
		}

	}

}
