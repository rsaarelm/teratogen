package sdl

/*
#include <SDL.h>
#include <SDL_mixer.h>
#include <SDL_ttf.h>

// Structs to help cgo handle some opaque SDL types.

typedef struct music { Mix_Music *music; } musicWrap;
typedef struct font { void *buf; TTF_Font *font; } fontWrap;

// XXX: Workaround to nested struct misalignment with cgo.
void unpackKeyeventKludge(SDL_KeyboardEvent* event, int* keysym, int* mod, int* unicode) {
 *keysym = event->keysym.sym;
 *mod = event->keysym.mod;
 *unicode = event->keysym.unicode;
}
*/
import "C"

import (
	"exp/draw"
	"fmt"
	"hyades/dbg"
	"hyades/keyboard"
	"hyades/sfx"
	"image"
	"os"
	"time"
	"unsafe"
)

const bitsPerPixel = 32

const keyBufferSize = 16

func withCString(str string, effect func(*C.char)) {
	cs := C.CString(str)
	effect(cs)
	C.free(unsafe.Pointer(cs))
}

//////////////////////////////////////////////////////////////////
// SDL Context object
//////////////////////////////////////////////////////////////////

type Context interface {
	draw.Context

	// SdlScreen is the same as Screen, but it return the screen cast into a
	// sdl.Surface instead of draw.Image.
	SdlScreen() Surface

	// Close Closes the SDL window and uninitializes SDL.
	Close()

	// Convert converts an image into a SDL surface
	Convert(img image.Image) Surface

	// TODO: API for freeing surfaces created by Convert. This should be done by a GC finalizer.

	// MakeSound converts wav file data into a SDL sound object.
	MakeSound(wavData []byte) (result sfx.Sound, err os.Error)

	// LoadMusic loads a piece of music from a file in the file system.
	// Due to limitations in SDL_Mixer, loading music from a byte buffer
	// isn't currently supported. Also, this seems to silenty fail with
	// MP3 files. OGG works. Music loops forever when played.
	LoadMusic(filename string) (result sfx.Sound, err os.Error)

	// KeyRepeatOn makes keyboard events repeat when a key is being held
	// down.
	KeyRepeatOn()

	// KeyRepeatOff makes a single keypress emit only a single keyboard
	// event no matter how long the key is held down.
	KeyRepeatOff()

	// LoadFont loads a font from TTF data.
	LoadFont(fontData []byte, pointSize int) (result Font, err os.Error)

	// Free frees the SDL resource the given value points to. If the value
	// doesn't point to a resource, does nothing.
	Free(handle interface{})
}

type Surface interface {
	draw.Image

	// Sets a clipping rectangle on a SDL surface. It's not possible to
	// draw outside the rectangle.
	SetClip(clipRect draw.Rectangle)

	// Clears a clipping rectangle on a SDL surface, if set.
	ClearClip()

	// Returns the clip rectangle of a SDL surface, if one has been set.
	GetClip() draw.Rectangle

	// Efficintly fills a rectangle on the screen with uniform color.
	FillRect(rect draw.Rectangle, c image.Color)

	// Blit draws an image on the surface. It is much more efficient if
	// the image is a SDL surface.
	Blit(img image.Image, x, y int)
}

type Font interface {
	Render(text string, color image.Color) (result image.Image, err os.Error)

	StringWidth(text string) int

	Height() int
}

type context struct {
	config Config

	screen *C.SDL_Surface

	kbd    chan int
	mouse  chan draw.Mouse
	resize chan bool
	quit   chan bool

	exitChan chan bool
}

type Config struct {
	Width      int
	Height     int
	Title      string
	Fullscreen bool
	Audio      bool
}

// NewWindow initializes SDL and returns a new SDL context.
func NewWindow(config Config) (result Context, err os.Error) {
	initFlags := int64(C.SDL_INIT_VIDEO)
	if config.Audio {
		initFlags |= C.SDL_INIT_AUDIO
	}
	screenFlags := int64(C.SDL_DOUBLEBUF)
	if config.Fullscreen {
		screenFlags |= C.SDL_FULLSCREEN
	}
	if C.SDL_Init(C.Uint32(initFlags)) == C.int(-1) {
		err = os.NewError(getError())
		return
	}
	screen := C.SDL_SetVideoMode(C.int(config.Width), C.int(config.Height), bitsPerPixel, C.Uint32(screenFlags))
	if screen == nil {
		err = os.NewError(getError())
		return
	}
	C.SDL_EnableUNICODE(1)
	if config.Audio {
		initAudio()
	}

	initTTF()

	ctx := new(context)
	result = ctx
	ctx.config = config
	ctx.screen = screen
	ctx.kbd = make(chan int, keyBufferSize)
	ctx.mouse = make(chan draw.Mouse, 1)
	ctx.resize = make(chan bool, 1)
	ctx.quit = make(chan bool, 1)
	ctx.exitChan = make(chan bool)

	go ctx.eventLoop()

	return
}

func (self *context) Screen() draw.Image { return self.screen }

func (self *context) SdlScreen() Surface { return self.screen }

func (self *context) FlushImage() { C.SDL_Flip(self.screen) }

func (self *context) KeyboardChan() <-chan int {
	return self.kbd
}

func (self *context) MouseChan() <-chan draw.Mouse {
	return self.mouse
}

func (self *context) ResizeChan() <-chan bool { return self.resize }

func (self *context) QuitChan() <-chan bool { return self.quit }

func (self *context) Close() {
	self.exitChan <- true
	// Wait for the event loop to finish and close SDL. The program may exit
	// without calling SDL_Quit if this isn't done.
	_ = <-self.exitChan
}

func (self *context) Convert(img image.Image) Surface {
	width, height := img.Width(), img.Height()

	var rmask, gmask, bmask, amask C.Uint32
	if C.SDL_BYTEORDER == C.SDL_BIG_ENDIAN {
		rmask, gmask, bmask, amask = 0xff000000, 0x00ff0000, 0x0000ff00, 0x000000ff
	} else {
		rmask, gmask, bmask, amask = 0x000000ff, 0x0000ff00, 0x00ff0000, 0xff000000
	}

	surf := C.SDL_CreateRGBSurface(0, C.int(width), C.int(height),
		C.int(self.screen.format.BitsPerPixel), rmask, gmask, bmask,
		amask)

	draw.Draw(surf, draw.Rect(0, 0, width, height), img, draw.Pt(0, 0))
	return surf
}

func (self *context) IsNativeSurface(img image.Image) bool {
	_, ok := img.(*C.SDL_Surface)
	return ok
}

func (self *context) MakeSound(wavData []byte) (result sfx.Sound, err os.Error) {
	if !self.config.Audio {
		err = os.NewError("Audio not active.")
		return
	}

	rw := C.SDL_RWFromMem(unsafe.Pointer(&wavData[0]), C.int(len(wavData)))
	chunk := C.Mix_LoadWAV_RW(rw, 1)
	if chunk == nil {
		err = os.NewError(getError())
		return
	}
	result = chunk
	return
}

func (self *context) LoadMusic(filename string) (result sfx.Sound, err os.Error) {
	if !self.config.Audio {
		err = os.NewError("Audio not active.")
		return
	}

	cs := C.CString(filename)
	music := &C.musicWrap{C.Mix_LoadMUS(cs)}
	C.free(unsafe.Pointer(cs))

	if music.music == nil {
		err = os.NewError(C.GoString(C.Mix_GetError()))
	}
	result = music
	return
}

func (self *context) KeyRepeatOn() {
	C.SDL_EnableKeyRepeat(C.SDL_DEFAULT_REPEAT_DELAY, C.SDL_DEFAULT_REPEAT_INTERVAL)
}

func (self *context) KeyRepeatOff() { C.SDL_EnableKeyRepeat(0, 0) }

func (self *context) eventLoop() {
	var evt C.SDL_Event

	const wheelUpBit = 1 << 3
	const wheelDownBit = 1 << 4

	for {
		if _, exit := <-self.exitChan; exit {
			break
		}
		if C.SDL_WaitEvent(&evt) != 0 {
			switch typ := eventType(&evt); typ {
			case C.SDL_KEYDOWN, C.SDL_KEYUP:
				keyEvt := ((*C.SDL_KeyboardEvent)(unsafe.Pointer(&evt)))
				self.handleKeyEvent(keyEvt, typ == C.SDL_KEYUP)
			case C.SDL_MOUSEMOTION:
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
			case C.SDL_MOUSEBUTTONDOWN, C.SDL_MOUSEBUTTONUP:
				btnEvt := ((*C.SDL_MouseButtonEvent)(unsafe.Pointer(&evt)))
				buttons := int(C.SDL_GetMouseState(nil, nil))
				if typ == C.SDL_MOUSEBUTTONDOWN && btnEvt.button == C.SDL_BUTTON_WHEELUP {
					buttons += wheelUpBit
				}
				if typ == C.SDL_MOUSEBUTTONDOWN && btnEvt.button == C.SDL_BUTTON_WHEELDOWN {
					buttons += wheelDownBit
				}
				mouse := draw.Mouse{buttons,
					draw.Pt(int(btnEvt.x), int(btnEvt.y)),
					time.Nanoseconds(),
				}
				_ = self.mouse <- mouse
			case C.SDL_VIDEORESIZE:
				_ = self.resize <- true
			case C.SDL_QUIT:
				_ = self.quit <- true
			}
		}
	}
	if self.config.Audio {
		exitAudio()
	}
	C.SDL_Quit()
	self.exitChan <- true
}

func (self *context) handleKeyEvent(keyEvt *C.SDL_KeyboardEvent, isKeyUp bool) {
	// Truncate unicode printable char to 16 bits, leave the rest of the bits
	// for special modifiers.

	// XXX: Commented out the code how this should work if cgo didn't have the
	// struct alignment bug.

	//chr := int(keyEvt.keysym.unicode) & 0xffff
	//sym := int(keyEvt.keysym.sym)
	// Key modifiers.
	//mod := int(keyEvt.keysym.mod)

	// XXX: cgo bug workaround
	var c_sym, c_mod, c_chr C.int
	C.unpackKeyeventKludge(keyEvt, &c_sym, &c_mod, &c_chr)
	sym := int(c_sym)
	mod := int(c_mod)
	chr := int(c_chr) & 0xffff
	isAscii := (sym >= 32 && sym < 127) || (sym > 128 && sym < 256)

	if isAscii && isKeyUp {
		// We don't get printable key information from SDL when raising pressed
		// keys. Good thing syms in the ascii range match printables.
		chr = sym
	}

	if !isAscii {
		// Nonprintable key.
		chr = keyboard.Nonprintable | sym
	}

	// XXX: Shift flag is *not* set for printable keys, to maintain the
	// convention that printable keys must provide printable char values.
	if !isAscii && mod&C.KMOD_LSHIFT != 0 {
		chr |= keyboard.LShift
	}
	if !isAscii && mod&C.KMOD_RSHIFT != 0 {
		chr |= keyboard.RShift
	}
	if mod&C.KMOD_LCTRL != 0 {
		chr |= keyboard.LCtrl
	}
	if mod&C.KMOD_RCTRL != 0 {
		chr |= keyboard.RCtrl
	}
	if mod&C.KMOD_LALT != 0 {
		chr |= keyboard.LAlt
	}
	if mod&C.KMOD_RALT != 0 {
		chr |= keyboard.RAlt
	}

	// As per the Context interface convention, key up is represented by a
	// negative key value.
	if isKeyUp {
		chr = -chr
	}

	// Non-blocking send.
	if ok := self.kbd <- chr; !ok {
		// Key buffer is full. Drop oldest key.
		_, _ = <-self.kbd
		_ = self.kbd <- chr
	}
}

func (self *context) Free(handle interface{}) {
	switch handle := handle.(type) {
	case (*C.SDL_Surface):
		C.SDL_FreeSurface(handle)
	case (*C.fontWrap):
		handle.Free()
	case (*C.musicWrap):
		C.Mix_FreeMusic(handle.music)
	case (*C.Mix_Chunk):
		C.Mix_FreeChunk(handle)
	default:
		fmt.Printf("Tried to free unknown resource type %v.\n", handle)
	}
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

// Cgo can't access the "type" field of the SDL_Event union type directly.
// This function casts the union into a minimal struct that contains only the
// leading type byte.
func eventType(evt *C.SDL_Event) byte {
	return byte(((*struct {
		_type C.Uint8
	})(unsafe.Pointer(evt)))._type)
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

	// XXX: Assuming 32-bit surface.
	pixels := (uintptr)(unsafe.Pointer(self.pixels))
	*(*uint32)(unsafe.Pointer(pixels + uintptr(y*int(self.pitch)+x<<2))) = color
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

func (self *C.SDL_Surface) Blit(img image.Image, x, y int) {
	if surf, ok := img.(*C.SDL_Surface); ok {
		// It's a SDL surface, do a fast SDL blit.
		rect := C.SDL_Rect{C.Sint16(x), C.Sint16(y), 0, 0}
		C.SDL_BlitSurface(surf, nil, self, &rect)
	} else {
		// It's something else, naively draw the individual pixels.
		draw.Draw(surf, draw.Rect(x, y, x+img.Width(), y+img.Height()),
			self, draw.Pt(0, 0))

	}
}

func (self *C.SDL_Surface) Width() int { return int(self.w) }

func (self *C.SDL_Surface) Height() int { return int(self.h) }

func (self *C.SDL_Surface) At(x, y int) image.Color {
	if !self.contains(x, y) {
		return image.RGBAColor{0, 0, 0, 0}
	}

	// XXX: Assuming 32-bit surface.
	pixels := (uintptr)(unsafe.Pointer(self.pixels))
	color := *(*uint32)(unsafe.Pointer(pixels + uintptr(y*int(self.pitch)+x<<2)))

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

func (self *C.SDL_Surface) SetClip(clipRect draw.Rectangle) {
	sdlClipRect := convertRect(clipRect)
	C.SDL_SetClipRect(self, &sdlClipRect)
}

func (self *C.SDL_Surface) ClearClip() { C.SDL_SetClipRect(self, nil) }

func (self *C.SDL_Surface) GetClip() draw.Rectangle {
	var sdlRect C.SDL_Rect
	C.SDL_GetClipRect(self, &sdlRect)
	return draw.Rect(int(sdlRect.x), int(sdlRect.y),
		int(sdlRect.x)+int(sdlRect.w), int(sdlRect.y)+int(sdlRect.h))
}

func sdlColor(color image.Color) C.SDL_Color {
	r, g, b, _ := color.RGBA()
	return C.SDL_Color{C.Uint8(r >> 24), C.Uint8(g >> 24), C.Uint8(b >> 24), 0}
}

//////////////////////////////////////////////////////////////////
// Audio
//////////////////////////////////////////////////////////////////

func initAudio() {
	var audioFormat C.Uint16
	switch sfx.DefaultSampleBytes {
	case sfx.Bit8:
		audioFormat = C.Uint16(C.AUDIO_S8)
	case sfx.Bit16:
		// XXX: Can't use AUDIO_S16 here as it's #defined to be "AUDIO_S16LSB",
		// and cgo doesn't chase #defines with non-literal values.
		audioFormat = C.Uint16(C.AUDIO_S16LSB)

	default:
		dbg.Die("Bad audioBytesPerSample %v", sfx.DefaultSampleBytes)
	}

	audioBuffers := C.int(4096)

	ok := C.Mix_OpenAudio(C.int(sfx.DefaultSampleRate), audioFormat, C.int(sfx.DefaultNumChannels), audioBuffers)

	if ok != 0 {
		panic("Mixer error: " + getError())
	}
}

func exitAudio() { C.Mix_CloseAudio() }

func (self *C.Mix_Chunk) Play() { C.Mix_PlayChannelTimed(-1, self, 0, -1) }

func (self *C.Mix_Chunk) Free() {
	if self != nil {
		C.Mix_FreeChunk(self)
	}
}

func (self *C.musicWrap) Play() { C.Mix_PlayMusic(self.music, -1) }

//////////////////////////////////////////////////////////////////
// TTF
//////////////////////////////////////////////////////////////////

func initTTF() {
	ok := C.TTF_Init()

	if ok != 0 {
		panic("TTF error: " + getError())
	}
}

func (self *context) LoadFont(fontData []byte, pointSize int) (result Font, err os.Error) {
	// XXX: Can't use the straight Go data since garbage collection would mess TTF.
	buf := C.malloc(C.size_t(len(fontData)))
	C.memcpy(buf, unsafe.Pointer(&fontData[0]), C.size_t(len(fontData)))
	rw := C.SDL_RWFromMem(buf, C.int(len(fontData)))
	//	rw := C.makeMemRwop(unsafe.Pointer(&fontData[0]), C.int(len(fontData)))

	if rw == nil {
		err = os.NewError(getError())
		C.free(buf)
		return
	}

	font := C.TTF_OpenFontRW(rw, 0, C.int(pointSize))

	if font == nil {
		err = os.NewError(getError())
		C.free(buf)
		return
	}

	// XXX: We need to hang on to the buf as well, since TTF fonts seem to stop
	// working properly if their RWops source data goes away.
	wrap := &C.fontWrap{buf, font}

	return wrap, err
}

func (self *C.fontWrap) Render(text string, color image.Color) (result image.Image, err os.Error) {
	var surface *C.SDL_Surface
	cs := C.CString(text)
	surface = C.TTF_RenderText_Solid(self.font, cs, sdlColor(color))
	C.free(unsafe.Pointer(cs))

	if surface == nil {
		err = os.NewError(getError())
		return
	}
	return surface, err
}

func (self *C.fontWrap) StringWidth(text string) int {
	var w C.int
	cs := C.CString(text)
	C.TTF_SizeText(self.font, cs, &w, nil)
	C.free(unsafe.Pointer(cs))
	return int(w)
}

func (self *C.fontWrap) Height() int { return int(C.TTF_FontHeight(self.font)) }

func (self *C.fontWrap) Free() {
	C.free(self.buf)
	C.TTF_CloseFont(self.font)
}
