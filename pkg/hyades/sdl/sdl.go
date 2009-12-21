package sdl

/*
#include <SDL.h>
#include <SDL_mixer.h>
*/
import "C"

import (
	"exp/draw"
	"hyades/dbg"
	"hyades/event"
	"image"
	"os"
	"time"
	"unsafe"
)

const bitsPerPixel = 32

//////////////////////////////////////////////////////////////////
// SDL Context object
//////////////////////////////////////////////////////////////////

type Sound interface {
	Play()
}

type Context interface {
	draw.Context

	// Close Closes the SDL window and uninitializes SDL.
	Close()

	// Blit draws an image on the window. It is much more efficient if the
	// image is implemented by a SDL surface.
	Blit(img image.Image, x, y int)

	// Efficintly fills a rectangle on the screen with uniform color.
	FillRect(rect draw.Rectangle, col image.Color)

	// Convert converts an image into a SDL surface
	Convert(img image.Image) image.Image

	// TODO: API for freeing surfaces created by Convert. This should be done by a GC finalizer.

	// MakeSound converts wav file data into a SDL sound object.
	MakeSound(wavData []byte) (result Sound, err os.Error)

	// KeyRepeatOn makes keyboard events repeat when a key is being held
	// down.
	KeyRepeatOn()

	// KeyRepeatOff makes a single keypress emit only a single keyboard
	// event no matter how long the key is held down.
	KeyRepeatOff()
}

type context struct {
	screen *C.SDL_Surface
	kbd    chan int
	mouse  chan draw.Mouse
	resize chan bool
	quit   chan bool

	active bool
}

// NewWindow initializes SDL and returns a new SDL context.
func NewWindow(width, height int, title string, fullscreen bool) (result Context, err os.Error) {
	flags := int64(DOUBLEBUF)
	if fullscreen {
		flags |= FULLSCREEN
	}
	if C.SDL_Init(INIT_VIDEO|INIT_AUDIO) == C.int(-1) {
		err = os.NewError(getError())
		return
	}
	screen := C.SDL_SetVideoMode(C.int(width), C.int(height), bitsPerPixel, C.Uint32(flags))
	if screen == nil {
		err = os.NewError(getError())
		return
	}
	C.SDL_EnableUNICODE(1)
	initAudio()

	ctx := new(context)
	result = ctx
	ctx.screen = screen
	ctx.kbd = make(chan int, 1)
	ctx.mouse = make(chan draw.Mouse, 1)
	ctx.resize = make(chan bool, 1)
	ctx.quit = make(chan bool, 1)
	ctx.active = true

	go ctx.eventLoop()

	return
}

func (self *context) Screen() draw.Image { return self.screen }

func (self *context) FlushImage() { C.SDL_Flip(self.screen) }

func (self *context) KeyboardChan() <-chan int {
	return self.kbd
}

func (self *context) MouseChan() <-chan draw.Mouse {
	return self.mouse
}

func (self *context) ResizeChan() <-chan bool { return self.resize }

func (self *context) QuitChan() <-chan bool { return self.quit }

func (self *context) Close() { self.active = false }

func (self *context) Blit(img image.Image, x, y int) {
	if surface, isSurface := img.(*C.SDL_Surface); isSurface {
		// It's a SDL surface, do a fast SDL blit.
		rect := C.SDL_Rect{C.Sint16(x), C.Sint16(y), 0, 0}
		C.SDL_BlitSurface(surface, nil, self.screen, &rect)
	} else {
		// It's something else, naively draw the individual pixels.
		draw.Draw(self.Screen(),
			draw.Rect(x, y, x+img.Width(), y+img.Height()), img, nil, draw.Pt(0, 0))
	}
}

func (self *context) FillRect(rect draw.Rectangle, c image.Color) {
	self.screen.FillRect(rect, c)
}

func (self *context) Convert(img image.Image) image.Image {
	width, height := img.Width(), img.Height()

	var rmask, gmask, bmask, amask C.Uint32
	if BYTEORDER == BIG_ENDIAN {
		rmask, gmask, bmask, amask = 0xff000000, 0x00ff0000, 0x0000ff00, 0x000000ff
	} else {
		rmask, gmask, bmask, amask = 0x000000ff, 0x0000ff00, 0x00ff0000, 0xff000000
	}

	surf := C.SDL_CreateRGBSurface(0, C.int(width), C.int(height),
		C.int(self.screen.format.BitsPerPixel), rmask, gmask, bmask,
		amask)

	draw.Draw(surf, draw.Rect(0, 0, width, height), img, nil, draw.Pt(0, 0))
	return surf
}

func (self *context) MakeSound(wavData []byte) (result Sound, err os.Error) {
	rw := C.SDL_RWFromMem(unsafe.Pointer(&wavData[0]), C.int(len(wavData)))
	chunk := C.Mix_LoadWAV_RW(rw, 1)
	if chunk == nil {
		err = os.NewError(getError())
		return
	}
	result = chunk
	return
}

func (self *context) KeyRepeatOn() {
	C.SDL_EnableKeyRepeat(DEFAULT_REPEAT_DELAY, DEFAULT_REPEAT_INTERVAL)
}

func (self *context) KeyRepeatOff() { C.SDL_EnableKeyRepeat(0, 0) }

func (self *context) eventLoop() {
	var evt C.SDL_Event

	const wheelUpBit = 1 << 3
	const wheelDownBit = 1 << 4

	for self.active {
		if C.SDL_PollEvent(&evt) != 0 {
			switch typ := eventType(&evt); typ {
			case KEYDOWN, KEYUP:
				keyEvt := ((*C.SDL_KeyboardEvent)(unsafe.Pointer(&evt)))

				// Truncate unicode printable char to 16 bits,
				// leave the rest of the bits for special
				// modifiers.
				chr := int(keyEvt.keysym.unicode) & 0xffff

				// TODO: Nonprintable special keys: SDL keysym value + bit 17 set.

				// TODO: Ctrl and Alt modifiers in high bits.

				// As per the Context interface, key up is
				// represented by a negative key value.
				if typ == KEYUP {
					chr = -chr
				}

				// Non-blocking send.
				_ = self.kbd <- chr
			case MOUSEMOTION:
				motEvt := ((*C.SDL_MouseMotionEvent)(unsafe.Pointer(&evt)))
				// XXX: SDL mouse button state *should* map
				// directly to draw.Mouse.Buttons. Still a bit
				// sloppy to just plug it in without a
				// converter...
				mouse := draw.Mouse{int(motEvt.state),
					draw.Pt(int(motEvt.x), int(motEvt.y)),
					time.Nanoseconds(),
				}
				// Non-blocking send
				_ = self.mouse <- mouse
			case MOUSEBUTTONDOWN, MOUSEBUTTONUP:
				btnEvt := ((*C.SDL_MouseButtonEvent)(unsafe.Pointer(&evt)))
				buttons := int(C.SDL_GetMouseState(nil, nil))
				if typ == MOUSEBUTTONDOWN && btnEvt.button == BUTTON_WHEELUP {
					buttons += wheelUpBit
				}
				if typ == MOUSEBUTTONDOWN && btnEvt.button == BUTTON_WHEELDOWN {
					buttons += wheelDownBit
				}
				mouse := draw.Mouse{buttons,
					draw.Pt(int(btnEvt.x), int(btnEvt.y)),
					time.Nanoseconds(),
				}
				_ = self.mouse <- mouse
			case VIDEORESIZE:
				_ = self.resize <- true
			case QUIT:
				_ = self.quit <- true
			}
		}
	}
	exitAudio()
	C.SDL_Quit()
}

//////////////////////////////////////////////////////////////////
// Helper functions
//////////////////////////////////////////////////////////////////

func getError() string { return C.GoString(C.SDL_GetError()) }

func convertRect(rect draw.Rectangle) C.SDL_Rect {
	rect = rect.Canon()
	return C.SDL_Rect{C.Sint16(rect.Min.X),
		C.Sint16(rect.Min.Y),
		C.Uint16(rect.Max.X - rect.Min.X),
		C.Uint16(rect.Max.Y - rect.Min.Y),
	}
}

// Due to syntax issues, I can't access a C struct field called "type"
// directly. This function implements an indirect way.
func eventType(evt *C.SDL_Event) byte {
	// XXX: Exploiting the fact that type is always the first field in the
	// struct. This isn't totally guaranteed, but I think SDL exploits it
	// too, so they're not that likely to change it.
	return *((*byte)(unsafe.Pointer(evt)))
}

//////////////////////////////////////////////////////////////////
// Video
//////////////////////////////////////////////////////////////////

func (self *C.SDL_Surface) FreeSurface() {
	if self != nil {
		C.SDL_FreeSurface(self)
	}
}

func (self *C.SDL_Surface) contains(x, y int) bool {
	return x < self.Width() && y < self.Height() && x >= 0 && y >= 0
}

func (self *C.SDL_Surface) Set(x, y int, c image.Color) {
	if !self.contains(x, y) {
		return
	}
	color := self.mapRGBA(c)

	// XXX: Calling another method here is pretty slow probably. Also
	// should unroll the loop for fixed ops for 1, 2, 3 and 4 bytes per
	// pixel.
	for i := 0; i < int(self.format.BytesPerPixel); i++ {
		self.writePixelData(self.pixelOffset(x, y)+i, byte(color%0x100))
		color = color >> 8
	}
}

func (self *C.SDL_Surface) FillRect(rect draw.Rectangle, c image.Color) {
	sdlRect := convertRect(rect)
	C.SDL_FillRect(self,
		&sdlRect,
		C.Uint32(self.mapRGBA(c)))
}

func (self *C.SDL_Surface) mapRGBA(c image.Color) uint32 {
	r32, g32, b32, a32 := c.RGBA()
	// TODO: Compensate for pre-alphamultiplication from c.RGBA(), intensify RGB if A is low.
	r, g, b, a := byte(r32>>24), byte(g32>>24), byte(b32>>24), byte(a32>>24)

	return uint32(C.SDL_MapRGBA(self.format,
		C.Uint8(r), C.Uint8(g), C.Uint8(b), C.Uint8(a)))
}

func (self *C.SDL_Surface) Blit(target *C.SDL_Surface, x, y int) {
	rect := C.SDL_Rect{C.Sint16(x), C.Sint16(y), 0, 0}
	C.SDL_BlitSurface(self, nil, target, &rect)
}

func (self *C.SDL_Surface) Width() int { return int(self.w) }

func (self *C.SDL_Surface) Height() int { return int(self.h) }

func (self *C.SDL_Surface) At(x, y int) image.Color {
	if !self.contains(x, y) {
		return image.RGBAColor{0, 0, 0, 0}
	}

	bitMask := uint32(0xffffffff) >> (32 - self.format.BitsPerPixel)
	color := self.readPixelData(self.pixelOffset(x, y)) & uint32(bitMask)
	var r, g, b, a byte
	C.SDL_GetRGBA(C.Uint32(color),
		self.format,
		(*C.Uint8)(&r), (*C.Uint8)(&g), (*C.Uint8)(&b), (*C.Uint8)(&a))
	return image.RGBAColor{r, g, b, a}
}

// For compliance wth the image.Image interface
func (self *C.SDL_Surface) ColorModel() image.ColorModel {
	return image.RGBAColorModel
}

func (self *C.SDL_Surface) writePixelData(offset int, data byte) {
	pixPtr := (uintptr)(unsafe.Pointer(self.pixels)) + uintptr(offset)
	*(*byte)(unsafe.Pointer(pixPtr)) = data
}

func (self *C.SDL_Surface) readPixelData(offset int) uint32 {
	pixPtr := (uintptr)(unsafe.Pointer(self.pixels)) + uintptr(offset)
	return *(*uint32)(unsafe.Pointer(pixPtr))
}

func (self *C.SDL_Surface) pixelOffset(x, y int) int {
	return y*int(self.pitch) + x*int(self.format.BytesPerPixel)
}

//////////////////////////////////////////////////////////////////
// Events
//////////////////////////////////////////////////////////////////

// Returns an event if there's one available, otherwise nil.
func PollEvent() event.Event {
	var evt C.SDL_Event
	if C.SDL_PollEvent(&evt) != 0 {
		return mapEvent(&evt)
	}
	return nil
}

func mapEvent(evt *C.SDL_Event) event.Event {
	if evt == nil {
		return nil
	}

	switch eventType(evt) {
	case KEYDOWN:
		keyEvt := ((*C.SDL_KeyboardEvent)(unsafe.Pointer(evt)))
		return &event.KeyDown{int(keyEvt.keysym.sym),
			int(keyEvt.keysym.unicode), uint(C.SDL_GetModState()),
		}
	case KEYUP:
		keyEvt := ((*C.SDL_KeyboardEvent)(unsafe.Pointer(evt)))
		return &event.KeyUp{int(keyEvt.keysym.sym),
			int(keyEvt.keysym.unicode), uint(C.SDL_GetModState()),
		}
	case MOUSEMOTION:
		motEvt := ((*C.SDL_MouseMotionEvent)(unsafe.Pointer(evt)))
		return &event.MouseMove{int(motEvt.x), int(motEvt.y),
			int(motEvt.xrel), int(motEvt.yrel), uint(motEvt.state), 0,
		}
	case MOUSEBUTTONDOWN:
		btnEvt := ((*C.SDL_MouseButtonEvent)(unsafe.Pointer(evt)))
		return &event.MouseDown{int(btnEvt.x), int(btnEvt.y), 0, 0,
			uint(C.SDL_GetMouseState(nil, nil)), int(btnEvt.button),
		}
	case MOUSEBUTTONUP:
		btnEvt := ((*C.SDL_MouseButtonEvent)(unsafe.Pointer(evt)))
		return &event.MouseUp{int(btnEvt.x), int(btnEvt.y), 0, 0,
			uint(C.SDL_GetMouseState(nil, nil)), int(btnEvt.button),
		}
	case VIDEORESIZE:
		resEvt := ((*C.SDL_ResizeEvent)(unsafe.Pointer(evt)))
		return &event.Resize{int(resEvt.w), int(resEvt.h)}
	case QUIT:
		result := new(event.Quit)
		return result
	}
	return nil
}


//////////////////////////////////////////////////////////////////
// Audio
//////////////////////////////////////////////////////////////////

// 8-bit, 4000 Hz audio
const audioRate = 4000
const audioBytesPerSample = 1
const audioChannels = 2

// XXX: Get rid of public AudioRateHz, AudioBytesPerSample, the outside
// shouldn't need to care about audio details.

// Return the preferred sample rate for audio samples.
func AudioRateHz() uint32 { return audioRate }

// Return the preferred byte count, 1 or 2, of audio samples
func AudioBytesPerSample() int { return audioBytesPerSample }

func initAudio() {
	var audioFormat C.Uint16
	switch audioBytesPerSample {
	case 1:
		audioFormat = C.Uint16(AUDIO_U8)
	case 2:
		audioFormat = C.Uint16(AUDIO_U16SYS)
	default:
		dbg.Die("Bad audioBytesPerSample %v", audioBytesPerSample)
	}

	audioBuffers := C.int(4096)

	ok := C.Mix_OpenAudio(C.int(audioRate), audioFormat, C.int(audioChannels), audioBuffers)

	if ok != 0 {
		panic("Mixer error" + getError())
	}
}

func exitAudio() { C.Mix_CloseAudio() }

func (self *C.Mix_Chunk) Play() { C.Mix_PlayChannelTimed(-1, self, C.int(0), -1) }

func (self *C.Mix_Chunk) Free() {
	if self != nil {
		C.Mix_FreeChunk(self)
	}
}
