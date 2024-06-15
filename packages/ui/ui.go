package ui

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	emu "pajalic.go.emulator/packages/cpu"
	log "pajalic.go.emulator/packages/logger"
	"unsafe"
)

const (
	SCREEN_WIDTH  = 640
	SCREEN_HEIGHT = 480
	scale         = 4
)

var (
	sdlWindow        *sdl.Window
	sdlRenderer      *sdl.Renderer
	sdlTexture       *sdl.Texture
	screen           *sdl.Surface
	sdlDebugWindow   *sdl.Window
	sdlDebugRenderer *sdl.Renderer
	sdlDebugTexture  *sdl.Texture
	debugScreen      *sdl.Surface
)

func UiInit() {
	log.Info("Cart loaded..")
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		log.Fatal("SDL init failed:", err)
	}
	log.Info("SDL INIT")

	if err := ttf.Init(); err != nil {
		log.Fatal("TTF init failed:", err)
	}
	log.Info("TTF INIT")

	var err error
	sdlWindow, sdlRenderer, err = sdl.CreateWindowAndRenderer(SCREEN_WIDTH, SCREEN_HEIGHT, sdl.WINDOW_RESIZABLE)
	if err != nil {
		log.Fatal("Failed to create window and renderer:", err)
	}

	sdlDebugWindow, sdlDebugRenderer, err = sdl.CreateWindowAndRenderer(16*8*scale, 32*8*scale, 0)
	if err != nil {
		log.Fatal("Failed to create debug window and renderer:", err)
	}

	debugScreen, err = sdl.CreateRGBSurface(0, (16*8*scale)+(16*scale),
		(32*8*scale)+(64*scale), 32,
		0x00FF0000,
		0x0000FF00,
		0x000000FF,
		0xFF000000)
	if err != nil {
		log.Fatal("Failed to create debug screen surface:", err)
	}

	sdlDebugTexture, err = sdlDebugRenderer.CreateTexture(
		sdl.PIXELFORMAT_ABGR8888,
		sdl.TEXTUREACCESS_STREAMING,
		(16*8*scale)+(16*scale),
		(32*8*scale)+(64*scale))
	if err != nil {
		log.Fatal("Failed to create debug texture:", err)
	}

	sdlTexture, err = sdlRenderer.CreateTexture(
		sdl.PIXELFORMAT_ARGB8888,
		sdl.TEXTUREACCESS_STREAMING,
		SCREEN_WIDTH,
		SCREEN_HEIGHT)
	if err != nil {
		log.Fatal("Failed to create SDL texture:", err)
	}

	screen, err = sdl.CreateRGBSurface(0, SCREEN_WIDTH, SCREEN_HEIGHT, 32,
		0x00FF0000,
		0x0000FF00,
		0x000000FF,
		0xFF000000)
	if err != nil {
		log.Fatal("Failed to create screen surface:", err)
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
	// Ensure surface is valid
	if surface == nil {
		log.Error("Invalid SDL surface provided.")
		return
	}

	var rc sdl.Rect

	for tileY := int32(0); tileY < 16; tileY += 2 {
		var b1 = emu.BusRead(startLocation + (tileNum * 16) + uint16(tileY))
		var b2 = emu.BusRead(startLocation + (tileNum * 16) + uint16(tileY) + 1)

		for bit := int32(7); bit >= 0; bit-- {
			var b1bit byte = 0
			var b2bit byte = 0
			if b1&(1<<bit) != 0 {
				b1bit = 1
			}
			if b2&(1<<bit) != 0 {
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
	var tileNum uint16 = 0
	var xDraw int32 = 0
	var yDraw int32 = 0
	var rc sdl.Rect
	rc.X = 0
	rc.Y = 0
	rc.W = debugScreen.W
	rc.H = debugScreen.H
	err := debugScreen.FillRect(&rc, 0xFF111111)
	if err != nil {
		log.Error(err.Error())
	}

	var addr uint16 = 0x8000

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

	err = sdlDebugTexture.Update(nil, unsafe.Pointer(&pixels[0]), int(debugScreen.Pitch))
	if err != nil {
		log.Error(err.Error())
	}
	err = sdlDebugRenderer.Clear()
	if err != nil {
		log.Error(err.Error())
	}
	err = sdlDebugRenderer.Copy(sdlDebugTexture, nil, nil)
	if err != nil {
		log.Error(err.Error())
	}

	sdlDebugRenderer.Present()
}

func UiUpdate() {
	// Ensure sdlRenderer and sdlTexture are valid
	if sdlRenderer == nil || sdlTexture == nil {
		log.Error("Invalid SDL renderer or texture.")
		return
	}

	var rc sdl.Rect
	rc.X = 0
	rc.Y = 0
	rc.W = 2048
	rc.H = 2048

	videoBuffer := emu.PpuCtx.VideoBuffer

	for lineNum := 0; lineNum < emu.YRES; lineNum++ {
		for x := 0; x < emu.XRES; x++ {
			rc.X = int32(x * scale)
			rc.Y = int32(lineNum * scale)
			rc.W = int32(scale)
			rc.H = int32(scale)

			screen.FillRect(&rc, videoBuffer[x+(lineNum*emu.XRES)])
		}
	}

	// Update SDL texture with updated screen pixels
	pixels := screen.Pixels()
	if err := sdlTexture.Update(nil, unsafe.Pointer(&pixels[0]), int(screen.Pitch)); err != nil {
		log.Error("Failed to update SDL texture:", err)
		return
	}

	// Render SDL texture
	sdlRenderer.Clear()
	sdlRenderer.Copy(sdlTexture, nil, nil)
	sdlRenderer.Present()

	// Update debug windows
	UpdateDbgWindows()
}

func UiHandleEvents() {
	var e sdl.Event

	for e = sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
		switch t := e.(type) {
		case *sdl.KeyboardEvent:
			if t.Type == sdl.KEYDOWN {
				uiOnKey(true, t.Keysym.Sym)
			} else if t.Type == sdl.KEYUP {
				uiOnKey(false, t.Keysym.Sym)
			}
		case *sdl.WindowEvent:
			if t.Event == sdl.WINDOWEVENT_CLOSE {
				//emu.GetEmuContext().Die = true
			}
		}
	}

}

func uiOnKey(down bool, keyCode sdl.Keycode) {
	state := emu.GamePadGetState()
	switch keyCode {
	case sdl.K_z:
		state.B = down
	case sdl.K_x:
		state.A = down
	case sdl.K_RETURN:
		state.Start = down
	case sdl.K_TAB:
		state.Select = down
	case sdl.K_UP:
		state.Up = down
	case sdl.K_DOWN:
		state.Down = down
	case sdl.K_LEFT:
		state.Left = down
	case sdl.K_RIGHT:
		state.Right = down
	}
}
