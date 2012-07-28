package sdl

/*
#cgo pkg-config: sdl
#include <SDL/SDL.h>
*/
import "C"

import (
	"image"
	"sync"
	"unsafe"
)

var mutex sync.Mutex

// Open starts a windowed SDL application.
func Open(width, height int) {
	mutex.Lock()
	defer mutex.Unlock()

	initFlags := int64(C.SDL_INIT_VIDEO)
	screenFlags := 0

	if C.SDL_Init(C.Uint32(initFlags)) == C.int(-1) {
		panic(getError())
	}

	screen := C.SDL_SetVideoMode(
		C.int(width), C.int(height), 32, C.Uint32(screenFlags))
	if screen == nil {
		panic(getError())
	}
	C.SDL_EnableUNICODE(1)
}

// Close exits the SDL application.
func Close() {
	mutex.Lock()
	defer mutex.Unlock()

	C.SDL_Quit()
}

func IsKeyDown(key KeySym) bool {
	var numKeys C.int
	keys := C.SDL_GetKeyState(&numKeys)

	if int(key) < 0 || int(key) >= int(numKeys) {
		return false
	}

	// Keys is a byte array
	// XXX: Is this idiomatic for accessing C arrays?
	result := (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(keys)) + uintptr(key)))

	return *result != 0
}

const DefaultRepeatDelay = C.SDL_DEFAULT_REPEAT_DELAY
const DefaultRepeatInterval = C.SDL_DEFAULT_REPEAT_INTERVAL

func EnableKeyRepeat(delay int, interval int) (ok bool) {
	return int(C.SDL_EnableKeyRepeat(C.int(delay), C.int(interval))) == 0
}

func getError() string {
	return C.GoString(C.SDL_GetError())
}

func convertRect(rect image.Rectangle) C.SDL_Rect {
	rect = rect.Canon()
	return C.SDL_Rect{
		C.Sint16(rect.Min.X),
		C.Sint16(rect.Min.Y),
		C.Uint16(rect.Max.X - rect.Min.X),
		C.Uint16(rect.Max.Y - rect.Min.Y),
	}
}
