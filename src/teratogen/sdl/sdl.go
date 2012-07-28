package sdl

/*
#cgo pkg-config: sdl
#include <SDL/SDL.h>
*/
import "C"

import (
	"image"
	"image/color"
	"reflect"
	"unsafe"
)

type MouseEvent struct {
	Pos     image.Point
	Buttons int8
}

type Surface struct {
	ptr *C.SDL_Surface
}

// Pixels returns a byte slice memory-mapped to the pixel buffer of the
// surface.
func (s *Surface) Pixels() (result []byte) {
	size := int(s.ptr.pitch) * int(s.ptr.h)

	result = []byte{}
	header := (*reflect.SliceHeader)(unsafe.Pointer(&result))
	*header = reflect.SliceHeader{uintptr(s.ptr.pixels), size, size}
	return
}

// Pixels32 is the same as Pixels, except that it returns an array of 32-bit
// values. This is convenient for pixel manipulation of 32-bit color surfaces.
func (s *Surface) Pixels32() (result []uint32) {
	size := int(s.ptr.pitch) * int(s.ptr.h) / 4

	result = []uint32{}
	header := (*reflect.SliceHeader)(unsafe.Pointer(&result))
	*header = reflect.SliceHeader{uintptr(s.ptr.pixels), size, size}
	return
}

func (s *Surface) Size() image.Point {
	return image.Pt(int(s.ptr.w), int(s.ptr.h))
}

// Pitch returns the byte span of a horizontal line in the surface.
func (s *Surface) Pitch() int {
	return int(s.ptr.pitch)
}

// Pitch32 returns the span of a horizontal line in the surface in 32-bit units.
func (s *Surface) Pitch32() int {
	pitch := int(s.ptr.pitch)
	if pitch%4 != 0 {
		panic("Pitch not divisible by 4")
	}
	return pitch / 4
}

// BPP returns the bytes per pixel of the surface.
func (s *Surface) BPP() int {
	return int(s.ptr.format.BytesPerPixel)
}

func (s *Surface) MapColor(c color.Color) uint32 {
	r, g, b, a := c.RGBA()
	r8, g8, b8, a8 := C.Uint8(r>>8), C.Uint8(g>>8), C.Uint8(b>>8), C.Uint8(a>>8)
	return uint32(C.SDL_MapRGBA(s.ptr.format, r8, g8, b8, a8))
}

func (s *Surface) GetColor(c32 uint32) color.Color {
	var r8, g8, b8, a8 C.Uint8
	C.SDL_GetRGBA(C.Uint32(c32), s.ptr.format, &r8, &g8, &b8, &a8)
	return color.RGBA{uint8(r8), uint8(g8), uint8(b8), uint8(a8)}
}

var kbd chan KeyEvent
var mouse chan MouseEvent
var quit chan bool

var exitChan chan bool

const keyBufferSize = 16

// Open starts a windowed SDL application.
func Open(width, height int) {
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

	kbd = make(chan KeyEvent, keyBufferSize)
	mouse = make(chan MouseEvent, 1)
	quit = make(chan bool, 1)
	exitChan = make(chan bool)

	go eventLoop()

	return
}

// KeyboardChan returns a channel which emits keyboard events from the SDL
// event loop.
func KeyboardChan() <-chan KeyEvent {
	return kbd
}

// MouseChan returns a channel which emits mouse events from the SDL
// event loop.
func MouseChan() <-chan MouseEvent {
	return mouse
}

// QuitChan returns a channel which emits close application events from the
// SDL event loop.
func QuitChan() <-chan bool { return quit }

// Close exits the SDL application.
func Close() {
	exitChan <- true
	// Wait for the event loop to finish and close SDL. The program may exit
	// without calling SDL_Quit if this isn't done.
	_ = <-exitChan
}

// Video returns the surface for the base SDL window.
func Video() *Surface {
	return &Surface{C.SDL_GetVideoSurface()}
}

// Flip swaps screen buffers with a double-buffered display mode. Use it to
// make the changes you made to the screen become visible.
func Flip() {
	C.SDL_Flip(C.SDL_GetVideoSurface())
}

func FillRect(rect image.Rectangle, color color.Color) {
	sdlRect := convertRect(rect)
	C.SDL_FillRect(C.SDL_GetVideoSurface(), &sdlRect, convertColor(color))
}

func Clear(color color.Color) {
	C.SDL_FillRect(C.SDL_GetVideoSurface(), nil, convertColor(color))
}

func IsKeyDown(key int) bool {
	var numKeys C.int
	keys := C.SDL_GetKeyState(&numKeys)

	if key < 0 || key >= int(numKeys) {
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

func eventLoop() {
	var evt C.SDL_Event

	defer func() { exitChan <- true }()
	defer C.SDL_Quit()

	for {
		select {
		case _ = <-exitChan:
			return
		default:
		}

		if C.SDL_WaitEvent(&evt) != 0 {
			switch typ := eventType(&evt); typ {
			case C.SDL_KEYDOWN, C.SDL_KEYUP:
				keyEvt := ((*C.SDL_KeyboardEvent)(unsafe.Pointer(&evt)))

				data := KeyEvent{
					rune(keyEvt.keysym.unicode),
					KeySym(keyEvt.keysym.sym),
					Scancode(keyEvt.keysym.scancode),
					typ == C.SDL_KEYUP}

				select {
				case kbd <- data: // Send new key
				case _ = <-kbd: // Buffer full, drop oldest key, then send
					kbd <- data
				default:
				}

			case C.SDL_MOUSEMOTION, C.SDL_MOUSEBUTTONUP, C.SDL_MOUSEBUTTONDOWN:
				motEvt := ((*C.SDL_MouseMotionEvent)(unsafe.Pointer(&evt)))
				data := MouseEvent{image.Pt(int(motEvt.x), int(motEvt.y)), int8(motEvt.state)}

				select {
				case mouse <- data:
				default:
				}
			case C.SDL_QUIT:
				quit <- true
			}
		}
	}
}

func getError() string {
	return C.GoString(C.SDL_GetError())
}

// Cgo can't access the "type" field of the SDL_Event union type directly
// because it collides with a Go reserved word. This function casts the union
// into a minimal struct that contains only the leading type byte and returns
// the byte.
func eventType(evt *C.SDL_Event) byte {
	return byte(((*struct {
		_type C.Uint8
	})(unsafe.Pointer(evt)))._type)
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

func convertColor(c color.Color) C.Uint32 {
	r, g, b, a := c.RGBA()
	r8, g8, b8, a8 := C.Uint8(r>>8), C.Uint8(g>>8), C.Uint8(b>>8), C.Uint8(a>>8)
	return C.SDL_MapRGBA(
		C.SDL_GetVideoSurface().format,
		r8, g8, b8, a8)
}
