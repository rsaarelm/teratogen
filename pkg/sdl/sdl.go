package sdl

/*
#include <SDL.h>

*/
import "C"

func InitSdl(width, height int, title string, fullscreen bool) {
	C.SDL_Init(INIT_VIDEO);
	C.SDL_SetVideoMode(C.int(width), C.int(height), C.int(32),
		DOUBLEBUF);
}

func ExitSdl() {
	C.SDL_Quit();
}

type SdlSurface struct {
	surf *C.SDL_Surface;
}

func (self *SdlSurface) freeSurface() {
	C.SDL_FreeSurface(self.surf);
}

//func (self *SdlSurface) make32BitSoftwareSurface(width, height int) {
//	self.freeSurface();
//	self.surf = C.SDL_CreateRGBSurface(
//}
