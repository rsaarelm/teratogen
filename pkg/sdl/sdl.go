package sdl

/*
#include <SDL.h>
int sdl_doublebuf = SDL_DOUBLEBUF;
int sdl_init_video = SDL_INIT_VIDEO;
*/
import "C"

func InitSdl(width, height int, title string, fullscreen bool) {
	C.SDL_Init(C.Uint32(C.sdl_init_video));
	C.SDL_SetVideoMode(C.int(width), C.int(height), C.int(32),
		C.Uint32(C.sdl_doublebuf));
}

func ExitSdl() {
	C.SDL_Quit();
}