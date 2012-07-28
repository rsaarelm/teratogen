package sdl

/*
#cgo pkg-config: sdl
#include <SDL/SDL.h>
*/
import "C"

import (
	"image"
	"time"
	"unsafe"
)

type KeyEvent struct {
	Print rune // Printable character
	Sym   KeySym
	Code  Scancode
	Mod   KeyMod
	KeyUp bool // True if this event was a key being unpressed
}

type MouseEvent struct {
	Pos     image.Point
	Buttons int8
}

type ResizeEvent image.Point

// True if focus was gained, false if it was lost.
type FocusEvent bool

type QuitEvent struct{}

type event struct {
	Type    uint8
	padding [23]byte
}

func (e *event) poll() bool {
	mutex.Lock()
	defer mutex.Unlock()

	return C.SDL_PollEvent((*C.SDL_Event)(unsafe.Pointer(e))) != 0
}

func (e *event) convert() interface{} {
	switch e.Type {
	case C.SDL_ACTIVEEVENT:
		aEvt := (*C.SDL_ActiveEvent)(unsafe.Pointer(e))
		return FocusEvent(aEvt.gain == 1)
	case C.SDL_KEYDOWN, C.SDL_KEYUP:
		keyEvt := (*C.SDL_KeyboardEvent)(unsafe.Pointer(e))
		return KeyEvent{
			rune(keyEvt.keysym.unicode),
			KeySym(keyEvt.keysym.sym),
			Scancode(keyEvt.keysym.scancode),
			KeyMod(keyEvt.keysym.mod),
			e.Type == C.SDL_KEYUP}
	case C.SDL_MOUSEMOTION, C.SDL_MOUSEBUTTONUP, C.SDL_MOUSEBUTTONDOWN:
		motEvt := ((*C.SDL_MouseMotionEvent)(unsafe.Pointer(e)))
		return MouseEvent{image.Pt(int(motEvt.x), int(motEvt.y)), int8(motEvt.state)}
	case C.SDL_VIDEORESIZE:
		rsEvt := ((*C.SDL_ResizeEvent)(unsafe.Pointer(e)))
		return ResizeEvent{int(rsEvt.w), int(rsEvt.h)}
	case C.SDL_QUIT:
		return QuitEvent{}
	}
	return nil
}

var Events = make(chan interface{})

func eventLoop() {
	e := &event{}
	for {
		if !e.poll() {
			time.Sleep(10 * 1e6)
			continue
		}
		if evt := e.convert(); evt != nil {
			Events <- evt
		}
	}
}

func init() {
	go eventLoop()
}
