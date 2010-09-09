package sdl

/*
#include <SDL.h>
#include <SDL_mixer.h>
#include <SDL_ttf.h>
#include <SDL_rotozoom.h>
#include <SDL_gfxPrimitives.h>

// Structs to help cgo handle some opaque SDL types.

typedef struct music { Mix_Music *music; } musicWrap;
typedef struct font { TTF_Font *font; } fontWrap;

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
	"hyades/geom"
	"hyades/gfx"
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

type sdlSurface C.SDL_Surface

type mixChunk C.Mix_Chunk

type musicWrap C.musicWrap

//////////////////////////////////////////////////////////////////
// SDL Context object
//////////////////////////////////////////////////////////////////

// XXX: The Context interface in exp/draw used to be like this, and the SDL
// Context was written against it. It was changed recently to use a single
// event channel instead of the multiple ones. I'm copying the old one here so
// I won't have to change apps that depend on the SDL context to change their
// channel use.
type oldExpDrawContext interface {
	Screen() draw.Image
	FlushImage()

	KeyboardChan() <-chan int
	MouseChan() <-chan draw.MouseEvent
	ResizeChan() <-chan bool
	QuitChan() <-chan bool
}

type Context interface {
	oldExpDrawContext

	// SdlScreen is the same as Screen, but it return the screen cast into a
	// sdl.Surface instead of draw.Image.
	SdlScreen() Surface

	// Close Closes the SDL window and uninitializes SDL.
	Close()

	// Convert converts an image into a SDL surface
	Convert(img image.Image) Surface

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
	SetClip(clipRect image.Rectangle)

	// Clears a clipping rectangle on a SDL surface, if set.
	ClearClip()

	// Returns the clip rectangle of a SDL surface, if one has been set.
	GetClip() image.Rectangle

	// Efficintly fills a rectangle on the screen with uniform color.
	FillRect(rect image.Rectangle, c image.Color)

	// Blit draws an image on the surface. It is much more efficient if
	// the image is a SDL surface.
	Blit(img image.Image, x, y int)

	// Zoom scales the given surface by horizontal sx and vertical sy into a
	// new surface.
	Zoom(sx, sy float64) Surface
}

type Font interface {
	Render(text string, color image.Color) (result image.Image, err os.Error)

	StringWidth(text string) int

	Height() int
}

type context struct {
	config Config

	canvas           *sdlSurface
	windowW, windowH int

	kbd    chan int
	mouse  chan draw.MouseEvent
	resize chan bool
	quit   chan bool

	exitChan chan bool
}

type Config struct {
	Width      int
	Height     int
	PixelScale int
	Title      string
	Fullscreen bool
	Audio      bool
}

const maxPixelScale = 32

// NewWindow initializes SDL and returns a new SDL context.
func NewWindow(config Config) (result Context, err os.Error) {
	if config.PixelScale < 1 {
		config.PixelScale = 1
	}
	if config.PixelScale > maxPixelScale {
		config.PixelScale = maxPixelScale
	}

	initFlags := int64(C.SDL_INIT_VIDEO)
	if config.Audio {
		initFlags |= C.SDL_INIT_AUDIO
	}
	// XXX: SDL_RESIZABLE flag seems to cause this to segfault on XMonad.
	screenFlags := int64(C.SDL_DOUBLEBUF)
	if config.Fullscreen {
		screenFlags |= C.SDL_FULLSCREEN
	}
	if C.SDL_Init(C.Uint32(initFlags)) == C.int(-1) {
		err = os.NewError(getError())
		return
	}
	screen := (*sdlSurface)(C.SDL_SetVideoMode(C.int(config.Width*config.PixelScale), C.int(config.Height*config.PixelScale), bitsPerPixel, C.Uint32(screenFlags)))
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
	ctx.canvas = ctx.createSurface(config.Width, config.Height)
	screenBounds := screen.Bounds()
	ctx.windowW, ctx.windowH = screenBounds.Max.X, screenBounds.Max.Y
	ctx.kbd = make(chan int, keyBufferSize)
	ctx.mouse = make(chan draw.MouseEvent, 1)
	ctx.resize = make(chan bool, 1)
	ctx.quit = make(chan bool, 1)
	ctx.exitChan = make(chan bool)

	go ctx.eventLoop()

	return
}

func (self *context) Screen() draw.Image { return self.canvas }

func (self *context) SdlScreen() Surface { return self.canvas }

func (self *context) FlushImage() {
	x, y := geom.CenterRects(
		self.config.Width*self.config.PixelScale,
		self.config.Height*self.config.PixelScale,
		self.windowW, self.windowH)
	zoomCanvas := self.canvas.Zoom(float64(self.config.PixelScale), float64(self.config.PixelScale))
	// Turn off alpha flags so that alpha components in the canvas won't cause
	// partial drawing to the screen.
	C.SDL_SetAlpha(zoomCanvas.(*sdlSurface).raw(), 0, 255)
	self.videoSurface().Blit(zoomCanvas, x, y)
	self.Free(zoomCanvas)
	C.SDL_Flip(self.videoSurface().raw())
}

func (self *context) KeyboardChan() <-chan int {
	return self.kbd
}

func (self *context) MouseChan() <-chan draw.MouseEvent {
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

func (self *context) videoSurface() *sdlSurface {
	return (*sdlSurface)(C.SDL_GetVideoSurface())
}

func (self *context) createSurface(width, height int) *sdlSurface {
	var rmask, gmask, bmask, amask C.Uint32
	if C.SDL_BYTEORDER == C.SDL_BIG_ENDIAN {
		rmask, gmask, bmask, amask = 0xff000000, 0x00ff0000, 0x0000ff00, 0x000000ff
	} else {
		rmask, gmask, bmask, amask = 0x000000ff, 0x0000ff00, 0x00ff0000, 0xff000000
	}

	return (*sdlSurface)(C.SDL_CreateRGBSurface(0, C.int(width), C.int(height),
		C.int(self.videoSurface().format.BitsPerPixel), rmask, gmask, bmask,
		amask))
}

func (self *context) Convert(img image.Image) Surface {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	surf := self.createSurface(width, height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			surf.Set(x, y, img.At(x, y))
		}
	}
	return surf
}

func (self *context) IsNativeSurface(img image.Image) bool {
	_, ok := img.(*sdlSurface)
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
	result = (*mixChunk)(chunk)
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
	result = (*musicWrap)(music)
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
				// directly to draw.MouseEvent.Buttons. Still a bit
				// sloppy to just plug it in without a
				// converter...
				mouse := draw.MouseEvent{int(motEvt.state),
					image.Pt(
						int(motEvt.x)/self.config.PixelScale,
						int(motEvt.y)/self.config.PixelScale),
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
				mouse := draw.MouseEvent{buttons,
					image.Pt(
						int(btnEvt.x)/self.config.PixelScale,
						int(btnEvt.y)/self.config.PixelScale),
					time.Nanoseconds(),
				}
				_ = self.mouse <- mouse
			case C.SDL_VIDEORESIZE:
				resizeEvt := ((*C.SDL_ResizeEvent)(unsafe.Pointer(&evt)))
				self.windowW, self.windowH = int(resizeEvt.w), int(resizeEvt.h)
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
	case (*sdlSurface):
		C.SDL_FreeSurface(handle.raw())
	case (*ttfFont):
		handle.Free()
	case (*musicWrap):
		C.Mix_FreeMusic(handle.music)
	case (*mixChunk):
		C.Mix_FreeChunk(handle.raw())
	default:
		fmt.Printf("Tried to free unknown resource type %v.\n", handle)
	}
}

//////////////////////////////////////////////////////////////////
// Helper functions
//////////////////////////////////////////////////////////////////

func getError() string { return C.GoString(C.SDL_GetError()) }

func convertRect(rect image.Rectangle) C.SDL_Rect {
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

func (self *sdlSurface) raw() *C.SDL_Surface {
	return (*C.SDL_Surface)(self)
}

func (self *sdlSurface) contains(x, y int) bool {
	return self.Bounds().Contains(image.Pt(x, y))
}

func (self *sdlSurface) Set(x, y int, c image.Color) {
	if !self.contains(x, y) {
		return
	}
	color := self.mapRGBA(c)

	// XXX: Assuming 32-bit surface.
	pixels := (uintptr)(unsafe.Pointer(self.pixels))
	*(*uint32)(unsafe.Pointer(pixels + uintptr(y*int(self.pitch)+x<<2))) = color
}

func (self *sdlSurface) FillRect(rect image.Rectangle, c image.Color) {
	r, g, b, a := gfx.RGBA8Bit(c)
	if a < 255 {
		// Alpha-blended rectangle.
		C.boxRGBA(self.raw(),
			C.Sint16(rect.Min.X), C.Sint16(rect.Min.Y),
			C.Sint16(rect.Max.X)-1, C.Sint16(rect.Max.Y)-1,
			C.Uint8(r), C.Uint8(g), C.Uint8(b), C.Uint8(a))
	} else {
		// Efficient FillRect rectangle if opaque alpha.
		sdlRect := convertRect(rect)
		C.SDL_FillRect(self.raw(),
			&sdlRect,
			C.Uint32(self.mapRGBA(c)))
	}
}

func (self *sdlSurface) mapRGBA(c image.Color) uint32 {
	// TODO: Compensate for pre-alphamultiplication from c.RGBA(), intensify RGB if A is low.
	r, g, b, a := gfx.RGBA8Bit(c)

	return uint32(C.SDL_MapRGBA(self.format,
		C.Uint8(r), C.Uint8(g), C.Uint8(b), C.Uint8(a)))
}

func (self *sdlSurface) Blit(img image.Image, x, y int) {
	if surf, ok := img.(*sdlSurface); ok {
		// It's a SDL surface, do a fast SDL blit.
		rect := C.SDL_Rect{C.Sint16(x), C.Sint16(y), 0, 0}
		C.SDL_BlitSurface(surf.raw(), nil, self.raw(), &rect)
	} else {
		// It's something else, naively draw the individual pixels.
		draw.Draw(surf, img.Bounds().Add(image.Pt(x, y)),
			self, image.Pt(0, 0))

	}
}

func (self *sdlSurface) Bounds() image.Rectangle {
	return image.Rect(0, 0, int(self.w), int(self.h))
}

func (self *sdlSurface) At(x, y int) image.Color {
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
func (self *sdlSurface) ColorModel() image.ColorModel {
	return image.RGBAColorModel
}

func (self *sdlSurface) SetClip(clipRect image.Rectangle) {
	sdlClipRect := convertRect(clipRect)
	C.SDL_SetClipRect(self.raw(), &sdlClipRect)
}

func (self *sdlSurface) ClearClip() { C.SDL_SetClipRect(self.raw(), nil) }

func (self *sdlSurface) GetClip() image.Rectangle {
	var sdlRect C.SDL_Rect
	C.SDL_GetClipRect(self.raw(), &sdlRect)
	return image.Rect(int(sdlRect.x), int(sdlRect.y),
		int(sdlRect.x)+int(sdlRect.w), int(sdlRect.y)+int(sdlRect.h))
}

func (self *sdlSurface) Zoom(sx, sy float64) Surface {
	return (*sdlSurface)(C.zoomSurface(self.raw(), C.double(sx), C.double(sy), 0))
}

func sdlColor(col image.Color) C.SDL_Color {
	r, g, b, _ := gfx.RGBA8Bit(col)
	return C.SDL_Color{C.Uint8(r), C.Uint8(g), C.Uint8(b), 0}
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

func (self *mixChunk) raw() *C.Mix_Chunk { return (*C.Mix_Chunk)(self) }

func (self *mixChunk) Play() { C.Mix_PlayChannelTimed(-1, self.raw(), 0, -1) }

func (self *musicWrap) Play() { C.Mix_PlayMusic(self.music, -1) }

//////////////////////////////////////////////////////////////////
// TTF
//////////////////////////////////////////////////////////////////

type ttfFont struct {
	wrap C.fontWrap
	data []byte
}

func initTTF() {
	ok := C.TTF_Init()

	if ok != 0 {
		panic("TTF error: " + getError())
	}
}

func (self *context) LoadFont(fontData []byte, pointSize int) (result Font, err os.Error) {
	// XXX: Apparently RWops TTF will fail if the underlying data buffer gets
	// garbage collected. So we'll need to hold on to the fontData buffer in
	// the Font structure.

	// XXX: Should copy fontData to a local array so that the caller can't
	// change it later since we need to keep it around.
	rw := C.SDL_RWFromMem(unsafe.Pointer(&fontData[0]), C.int(len(fontData)))

	if rw == nil {
		err = os.NewError(getError())
		return
	}

	font := C.TTF_OpenFontRW(rw, 1, C.int(pointSize))

	if font == nil {
		err = os.NewError(getError())
		return
	}

	result = &ttfFont{C.fontWrap{font}, fontData}

	return
}

func (self *ttfFont) Render(text string, color image.Color) (result image.Image, err os.Error) {
	var surface *C.SDL_Surface
	cs := C.CString(text)
	surface = C.TTF_RenderText_Solid(self.wrap.font, cs, sdlColor(color))
	C.free(unsafe.Pointer(cs))

	if surface == nil {
		err = os.NewError(getError())
		return
	}
	return (*sdlSurface)(surface), err
}

func (self *ttfFont) StringWidth(text string) int {
	var w C.int
	cs := C.CString(text)
	C.TTF_SizeText(self.wrap.font, cs, &w, nil)
	C.free(unsafe.Pointer(cs))
	return int(w)
}

func (self *ttfFont) Height() int { return int(C.TTF_FontHeight(self.wrap.font)) }

func (self *ttfFont) Free() { C.TTF_CloseFont(self.wrap.font) }
