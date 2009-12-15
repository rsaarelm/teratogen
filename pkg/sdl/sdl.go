package sdl

/*
#include <SDL.h>

*/
import "C"

import (
	"io"
	"hyades/event"
	"image"
	"image/png"
	"os"
	"unsafe"
)

func InitSdl(width, height int, title string, fullscreen bool) {
	C.SDL_Init(INIT_VIDEO)
	C.SDL_SetVideoMode(C.int(width), C.int(height), 32, DOUBLEBUF)
	C.SDL_EnableUNICODE(1)
}

func ExitSdl()	{ C.SDL_Quit() }

type IntRect interface {
	X() int
	Y() int
	Width() int
	Height() int
}

// Structurally equivalent to SDL_Rect
type rect struct {
	x, y int16
	w, h uint16
}

func Rect(x, y int16, width, height uint16) IntRect {
	return &rect{x, y, width, height}
}

func (self *rect) X() int { return int(self.x) }

func (self *rect) Y() int { return int(self.y) }

func (self *rect) Width() int { return int(self.w) }

func (self *rect) Height() int { return int(self.h) }

func convertRect(rec IntRect) *rect {
	return &rect{int16(rec.X()), int16(rec.Y()),
		uint16(rec.Width()), uint16(rec.Height())}
}

func GetError() string {
	return C.GoString(C.SDL_GetError())
}

//////////////////////////////////////////////////////////////////
// Video
//////////////////////////////////////////////////////////////////

func Flip() {
	C.SDL_Flip(C.SDL_GetVideoSurface())
}

type Surface struct {
	surf *surface
}

func GetVideoSurface() (result *Surface) {
	result = new(Surface)
	result.surf = (*surface)(unsafe.Pointer(C.SDL_GetVideoSurface()))
	return
}

func Make32BitSurface(flags int, width, height int) (result *Surface) {
	result = new(Surface)

	var rmask, gmask, bmask, amask uint32

	if BYTEORDER == BIG_ENDIAN {
		rmask, gmask, bmask, amask = 0xff000000, 0x00ff0000, 0x0000ff00, 0x000000ff
	} else {
		rmask, gmask, bmask, amask = 0x000000ff, 0x0000ff00, 0x00ff0000, 0xff000000
	}

	result.surf = (*surface)(unsafe.Pointer(C.SDL_CreateRGBSurface(
		C.Uint32(flags), C.int(width), C.int(height), 32,
		C.Uint32(rmask), C.Uint32(gmask), C.Uint32(bmask), C.Uint32(amask))))
	return
}

func MakePngSurface(input io.Reader) (result *Surface, err os.Error) {
	pic, err := png.Decode(input)

	if err != nil {
		return nil, err
	}

	return MakeImageSurface(pic), nil
}

func MakeImageSurface(img image.Image) (result *Surface) {
	result = Make32BitSurface(0, img.Width(), img.Height())
	for x, w := 0, img.Width(); x < w; x++ {
		for y, h := 0, img.Height(); y < h; y++ {
			result.Set(x, y, img.At(x, y))
		}
	}
	return result
}

func (self *Surface) FreeSurface() {
	if self.surf != nil {
		C.SDL_FreeSurface((*C.SDL_Surface)(unsafe.Pointer(self.surf)))
		self.surf = nil
	}
}

func (self *Surface) Set(x, y int, c image.Color) {
	color := self.mapRGBA(c)

	// XXX: Calling another method here is pretty slow probably. Also
	// should unroll the loop for fixed ops for 1, 2, 3 and 4 bytes per
	// pixel.
	for i := 0; i < int(self.surf.Format.BytesPerPixel); i++ {
		self.writePixelData(self.pixelOffset(x, y) + i, byte(color % 0x100))
		color = color >> 8
	}
}

func (self *Surface) FillRect(rec IntRect, c image.Color) {
	C.SDL_FillRect((*C.SDL_Surface)(unsafe.Pointer(self.surf)),
		(*C.SDL_Rect)(unsafe.Pointer(convertRect(rec))),
		C.Uint32(self.mapRGBA(c)))
}

func (self *Surface) mapRGBA(c image.Color) uint32 {
	r32, g32, b32, a32 := c.RGBA()
	// TODO: Compensate for pre-alphamultiplication from c.RGBA(), intensify RGB if A is low.
	r, g, b, a := byte(r32 >> 24), byte(g32 >> 24), byte(b32 >> 24), byte(a32 >> 24)

	return uint32(C.SDL_MapRGBA((*C.SDL_PixelFormat)(unsafe.Pointer(self.surf.Format)),
		C.Uint8(r), C.Uint8(g), C.Uint8(b), C.Uint8(a)))
}

func (self *Surface) Blit(target *Surface, x, y int) {
	var srcRect, dstRect *rect
	dstRect = &rect{int16(x), int16(y), 0, 0}
	C.SDL_BlitSurface(
		(*C.SDL_Surface)(unsafe.Pointer(self.surf)),
		(*C.SDL_Rect)(unsafe.Pointer(srcRect)),
		(*C.SDL_Surface)(unsafe.Pointer(target.surf)),
		(*C.SDL_Rect)(unsafe.Pointer(dstRect)))
}

func (self *Surface) BlitRect(target *Surface, area IntRect, x, y int) {
	var srcRect, dstRect *rect
	srcRect = convertRect(area)
	dstRect = &rect{int16(x), int16(y), 0, 0}
	C.SDL_BlitSurface(
		(*C.SDL_Surface)(unsafe.Pointer(self.surf)),
		(*C.SDL_Rect)(unsafe.Pointer(srcRect)),
		(*C.SDL_Surface)(unsafe.Pointer(target.surf)),
		(*C.SDL_Rect)(unsafe.Pointer(dstRect)))
}

func (self *Surface) Width() int { return int(self.surf.W) }

func (self *Surface) Height() int { return int(self.surf.H) }

func (self *Surface) At(x, y int) image.Color {
	bitMask := uint32(0xffffffff) >> (32 - self.surf.Format.BitsPerPixel)
	color := self.readPixelData(self.pixelOffset(x, y)) & uint32(bitMask)
	var r, g, b, a byte
	C.SDL_GetRGBA(C.Uint32(color),
		(*C.SDL_PixelFormat)(unsafe.Pointer(self.surf.Format)),
		(*C.Uint8)(&r), (*C.Uint8)(&g), (*C.Uint8)(&b), (*C.Uint8)(&a))
	return image.RGBAColor{r, g, b, a}
}

// For compliance wth the image.Image interface
func (self *Surface) ColorModel() image.ColorModel {
	return image.RGBAColorModel
}

// Convert the pixel format of this surface to match that of the other one.
// Converting sprite surfaces to match the format of the display surface makes
// blitting them much faster.
func (self *Surface) Convert(other *Surface) {
	// TODO: More graceful error handling.
	newSurface := C.SDL_ConvertSurface(
		(*C.SDL_Surface)(unsafe.Pointer(self.surf)),
		(*C.SDL_PixelFormat)(unsafe.Pointer(other.surf.Format)),
		C.Uint32(self.surf.Flags))
	if newSurface == nil {
		panic("Surface conversion failed")
	}
	self.FreeSurface()
	self.surf = (*surface)(unsafe.Pointer(newSurface))
}

func (self *Surface) MakeTiles(width, height int,
	offsetX, offsetY int,
	gapX, gapY int) (result []*Surface) {
	numX := (self.Width() - offsetX) / (width + gapX)
	numY := (self.Height() - offsetY) / (width + gapY)

	result = make([]*Surface, numX * numY)
	i := 0

	for x := 0; x < numX; x++ {
		for y := 0; y < numY; y++ {
			tile := Make32BitSurface(int(self.surf.Flags), width, height)
			rect := Rect(int16(offsetX + x * (width + gapX)),
				int16(offsetY + y * (height + gapY)),
				uint16(width), uint16(height))

			tile.Convert(self)
			self.BlitRect(tile, rect, 0, 0)

			result[i] = tile
			i++
		}
	}
	return
}

func (self *Surface) writePixelData(offset int, data byte) {
	pixPtr := (uintptr)(unsafe.Pointer(self.surf.Pixels)) + uintptr(offset)
	*(*byte)(unsafe.Pointer(pixPtr)) = data
}

func (self *Surface) readPixelData(offset int) uint32 {
	pixPtr := (uintptr)(unsafe.Pointer(self.surf.Pixels)) + uintptr(offset)
	return *(*uint32)(unsafe.Pointer(pixPtr))
}

func (self *Surface) pixelOffset(x, y int) int {
	return y * int(self.surf.Pitch) + x * int(self.surf.Format.BytesPerPixel)
}

func (self *Surface) mustLock() bool {
	// Reimplement this macro from SDL_video.h:
	//#define SDL_MUSTLOCK(surface)   \
	//  (surface->offset ||           \
	//  ((surface->flags & (SDL_HWSURFACE|SDL_ASYNCBLIT|SDL_RLEACCEL)) != 0))
	return self.surf.Offset != 0 || self.surf.Flags & (HWSURFACE|ASYNCBLIT|RLEACCEL) != 0
}

//////////////////////////////////////////////////////////////////
// Events
//////////////////////////////////////////////////////////////////

func mapEvent(evt *C.SDL_Event) event.Event {
	if evt == nil { return nil }

	switch ((*_event)(unsafe.Pointer(evt))).Type {
	case KEYDOWN:
		keyEvt := ((*keyboardEvent)(unsafe.Pointer(evt)))
		sym := ((*keysym)(unsafe.Pointer(&keyEvt.Keysym)))
		return &event.KeyDown{int(sym.Sym),
			int(sym.Unicode), uint(C.SDL_GetModState())}
	case KEYUP:
		keyEvt := ((*keyboardEvent)(unsafe.Pointer(evt)))
		sym := ((*keysym)(unsafe.Pointer(&keyEvt.Keysym)))
		return &event.KeyUp{int(sym.Sym),
			int(sym.Unicode), uint(C.SDL_GetModState())}
	case MOUSEMOTION:
		motEvt := ((*mouseMotionEvent)(unsafe.Pointer(evt)))
		return &event.MouseMove{int(motEvt.X), int(motEvt.Y), int(motEvt.Xrel), int(motEvt.Yrel),
			uint(motEvt.State), 0}
	case MOUSEBUTTONDOWN:
		btnEvt := ((*mouseButtonEvent)(unsafe.Pointer(evt)))
		return &event.MouseDown{int(btnEvt.X), int(btnEvt.Y), 0, 0,
			uint(C.SDL_GetMouseState(nil, nil)), int(btnEvt.Button)}
	case MOUSEBUTTONUP:
		btnEvt := ((*mouseButtonEvent)(unsafe.Pointer(evt)))
		return &event.MouseUp{int(btnEvt.X), int(btnEvt.Y), 0, 0,
			uint(C.SDL_GetMouseState(nil, nil)), int(btnEvt.Button)}
	case VIDEORESIZE:
		resEvt := ((*resizeEvent)(unsafe.Pointer(evt)))
		return &event.Resize{int(resEvt.W), int(resEvt.H)}
	case QUIT:
		result := new(event.Quit)
		return result
	}
	return nil
}

// Loops forever waiting for SDL events, converting them to Hyades events and
// pushing them into the channel. Run as goroutine.
func EventListener(ch chan<- event.Event) {
	var evt C.SDL_Event
	for {
		err := C.SDL_WaitEvent(&evt)
		if err == 0 { continue /* TODO: Error handling. */ }
		localEvt := mapEvent(&evt)
		if localEvt != nil {
			ch <- localEvt
		}
	}
}

func KeyRepeatOn() {
	C.SDL_EnableKeyRepeat(DEFAULT_REPEAT_DELAY, DEFAULT_REPEAT_INTERVAL)
}

func KeyRepeatOff() {
	C.SDL_EnableKeyRepeat(0, 0)
}

// Written out manually since godefs gets the layout wrong.
type keysym struct {
	Scancode uint8
	Sym uint32
	Mod uint32
	Unicode uint16
}
