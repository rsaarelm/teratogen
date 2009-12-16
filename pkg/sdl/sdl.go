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
	flags := int64(DOUBLEBUF)
	if fullscreen { flags |= FULLSCREEN }
	C.SDL_Init(INIT_VIDEO)
	C.SDL_SetVideoMode(C.int(width), C.int(height), 32, C.Uint32(flags))
	C.SDL_EnableUNICODE(1)
}

func ExitSdl()	{ C.SDL_Quit() }

type IntRect interface {
	X() int
	Y() int
	Width() int
	Height() int
}

func Rect(x, y int16, width, height uint16) IntRect {
	return &C.SDL_Rect{C.Sint16(x), C.Sint16(y), C.Uint16(width), C.Uint16(height)}
}

func (self *C.SDL_Rect) X() int { return int(self.x) }

func (self *C.SDL_Rect) Y() int { return int(self.y) }

func (self *C.SDL_Rect) Width() int { return int(self.w) }

func (self *C.SDL_Rect) Height() int { return int(self.h) }

func convertRect(rec IntRect) *C.SDL_Rect {
	return &C.SDL_Rect{C.Sint16(rec.X()), C.Sint16(rec.Y()),
		C.Uint16(rec.Width()), C.Uint16(rec.Height())}
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
	surf *C.SDL_Surface
	// The area of the surface that'll be blitted. Assume entire surface
	// if nil.
	blitRect *C.SDL_Rect
}

func GetVideoSurface() (result *Surface) {
	// XXX: This is pretty immutable, could be cached?
	result = new(Surface)
	result.surf = C.SDL_GetVideoSurface()
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

	result.surf = C.SDL_CreateRGBSurface(
		C.Uint32(flags), C.int(width), C.int(height), 32,
		C.Uint32(rmask), C.Uint32(gmask), C.Uint32(bmask), C.Uint32(amask))
	// XXX: Need to init all to opaque alpha or blits won't set alpha.
	result.FillRect(Rect(0, 0, uint16(result.Width()), uint16(result.Height())),
		image.RGBAColor{0, 0, 0, 255})
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
		C.SDL_FreeSurface(self.surf)
		self.surf = nil
	}
}

func (self *Surface) Set(x, y int, c image.Color) {
	color := self.mapRGBA(c)

	// XXX: Calling another method here is pretty slow probably. Also
	// should unroll the loop for fixed ops for 1, 2, 3 and 4 bytes per
	// pixel.
	for i := 0; i < int(self.surf.format.BytesPerPixel); i++ {
		self.writePixelData(self.pixelOffset(x, y) + i, byte(color % 0x100))
		color = color >> 8
	}
}

func (self *Surface) FillRect(rec IntRect, c image.Color) {
	C.SDL_FillRect(self.surf,
	(*C.SDL_Rect)(unsafe.Pointer(convertRect(rec))),
	C.Uint32(self.mapRGBA(c))) }

func (self *Surface) mapRGBA(c image.Color) uint32 {
	r32, g32, b32, a32 := c.RGBA()
	// TODO: Compensate for pre-alphamultiplication from c.RGBA(), intensify RGB if A is low.
	r, g, b, a := byte(r32 >> 24), byte(g32 >> 24), byte(b32 >> 24), byte(a32 >> 24)

	return uint32(C.SDL_MapRGBA(self.surf.format,
		C.Uint8(r), C.Uint8(g), C.Uint8(b), C.Uint8(a)))
}

func (self *Surface) Blit(target *Surface, x, y int) {
	rect := C.SDL_Rect{C.Sint16(x), C.Sint16(y), 0, 0}
	C.SDL_BlitSurface(self.surf, self.blitRect, target.surf, &rect)
}

func (self *Surface) BlitRect(target *Surface, area IntRect, x, y int) {
	rect := C.SDL_Rect{C.Sint16(x), C.Sint16(y), 0, 0}
	srcRect := convertRect(area)
	C.SDL_BlitSurface(self.surf, srcRect, target.surf, &rect)
}

func (self *Surface) Width() int { return int(self.surf.w) }

func (self *Surface) Height() int { return int(self.surf.h) }

func (self *Surface) At(x, y int) image.Color {
	bitMask := uint32(0xffffffff) >> (32 - self.surf.format.BitsPerPixel)
	color := self.readPixelData(self.pixelOffset(x, y)) & uint32(bitMask)
	var r, g, b, a byte
	C.SDL_GetRGBA(C.Uint32(color),
		self.surf.format,
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
	newSurface := C.SDL_ConvertSurface(self.surf, other.surf.format,
		C.Uint32(self.surf.flags))
	if newSurface == nil {
		panic("Surface conversion failed")
	}
	self.FreeSurface()
	self.surf = newSurface
}

func (self *Surface) MakeTiles(width, height int,
	offsetX, offsetY int,
	gapX, gapY int) (result []*Surface) {
	numX := (self.Width() - offsetX) / (width + gapX)
	numY := (self.Height() - offsetY) / (width + gapY)

	result = make([]*Surface, numX * numY)
	i := 0

	for y := 0; y < numY; y++ {
		for x := 0; x < numX; x++ {
			tile := Make32BitSurface(int(self.surf.flags)|SRCALPHA, width, height)
			rect := Rect(int16(offsetX + x * (width + gapX)),
				int16(offsetY + y * (height + gapY)),
				uint16(width), uint16(height))

			self.Blit(tile, -rect.X(), -rect.Y())
			tile.Convert(self)

			result[i] = tile
			i++
		}
	}
	return
}

func (self *Surface) writePixelData(offset int, data byte) {
	pixPtr := (uintptr)(unsafe.Pointer(self.surf.pixels)) + uintptr(offset)
	*(*byte)(unsafe.Pointer(pixPtr)) = data
}

func (self *Surface) readPixelData(offset int) uint32 {
	pixPtr := (uintptr)(unsafe.Pointer(self.surf.pixels)) + uintptr(offset)
	return *(*uint32)(unsafe.Pointer(pixPtr))
}

func (self *Surface) pixelOffset(x, y int) int {
	return y * int(self.surf.pitch) + x * int(self.surf.format.BytesPerPixel)
}

func (self *Surface) mustLock() bool {
	// Reimplement this macro from SDL_video.h:
	//#define SDL_MUSTLOCK(surface)   \
	//  (surface->offset ||           \
	//  ((surface->flags & (SDL_HWSURFACE|SDL_ASYNCBLIT|SDL_RLEACCEL)) != 0))
	return self.surf.offset != 0 || self.surf.flags & (HWSURFACE|ASYNCBLIT|RLEACCEL) != 0
}

//////////////////////////////////////////////////////////////////
// Events
//////////////////////////////////////////////////////////////////

// Loops forever waiting for SDL events, converting them to Hyades events and
// pushing them into the channel. Run as goroutine.
func EventListener(ch chan<- event.Event) {
	var evt C.SDL_Event
	for {
		err := C.SDL_WaitEvent(&evt)
		if err == 0 {
			// TODO: Error handling
			continue
		}
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

func mapEvent(evt *C.SDL_Event) event.Event {
	if evt == nil { return nil }

	switch eventType(evt) {
	case KEYDOWN:
		keyEvt := ((*C.SDL_KeyboardEvent)(unsafe.Pointer(evt)))
		return &event.KeyDown{int(keyEvt.keysym.sym),
			int(keyEvt.keysym.unicode), uint(C.SDL_GetModState())}
	case KEYUP:
		keyEvt := ((*C.SDL_KeyboardEvent)(unsafe.Pointer(evt)))
		return &event.KeyUp{int(keyEvt.keysym.sym),
			int(keyEvt.keysym.unicode), uint(C.SDL_GetModState())}
	case MOUSEMOTION:
		motEvt := ((*C.SDL_MouseMotionEvent)(unsafe.Pointer(evt)))
		return &event.MouseMove{int(motEvt.x), int(motEvt.y),
			int(motEvt.xrel), int(motEvt.yrel), uint(motEvt.state), 0}
	case MOUSEBUTTONDOWN:
		btnEvt := ((*C.SDL_MouseButtonEvent)(unsafe.Pointer(evt)))
		return &event.MouseDown{int(btnEvt.x), int(btnEvt.y), 0, 0,
			uint(C.SDL_GetMouseState(nil, nil)), int(btnEvt.button)}
	case MOUSEBUTTONUP:
		btnEvt := ((*C.SDL_MouseButtonEvent)(unsafe.Pointer(evt)))
		return &event.MouseUp{int(btnEvt.x), int(btnEvt.y), 0, 0,
			uint(C.SDL_GetMouseState(nil, nil)), int(btnEvt.button)}
	case VIDEORESIZE:
		resEvt := ((*C.SDL_ResizeEvent)(unsafe.Pointer(evt)))
		return &event.Resize{int(resEvt.w), int(resEvt.h)}
	case QUIT:
		result := new(event.Quit)
		return result
	}
	return nil
}

// Due to syntax issues, I can't access a C struct field called "type"
// directly. This function implements an indirect way.
func eventType(evt *C.SDL_Event) byte {
	// XXX: Exploiting the fact that type is always the first field in the
	// struct. This isn't totally guaranteed, but I think SDL exploits it
	// too, so they're not that likely to change it.
	return *((*byte)(unsafe.Pointer(evt)))
}
