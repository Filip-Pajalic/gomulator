package ui

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	emu "pajalic.go.emulator/packages/emulator"
	log "pajalic.go.emulator/packages/logger"
)

var SCREEN_WIDTH int32 = 1024

var SCREEN_HEIGHT int32 = 768

var SDL_Window *sdl.Window

var SDL_Renderer *sdl.Renderer

var SDL_Texture *sdl.Texture

var SDL_Surface *sdl.Surface

func UiInit() {

	log.Info("Cart loaded..")
	sdl.Init(sdl.INIT_VIDEO)
	log.Info("SDL INIT")
	ttf.Init()
	log.Info("TTF INIT")
	var err error
	SDL_Window, SDL_Renderer, err = sdl.CreateWindowAndRenderer(SCREEN_WIDTH, SCREEN_HEIGHT, sdl.WINDOW_RESIZABLE)
	if err != nil {
		log.Fatal(err.Error())
	}

}

func DestroyWindow() {
	SDL_Window.Destroy()
}

func delay(ms uint32) {
	sdl.Delay(ms)
}

func UiHandleEvents() {

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
