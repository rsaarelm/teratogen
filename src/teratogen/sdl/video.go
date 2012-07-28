package sdl

/*
#include <SDL/SDL.h>
*/
import "C"

import (
	"image"
	"image/color"
	"image/draw"
	"reflect"
	"runtime"
	"unsafe"
)

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

func (s *Surface) ColorModel() color.Model {
	return color.RGBAModel
}

func (s *Surface) Bounds() image.Rectangle {
	return image.Rect(0, 0, int(s.ptr.w), int(s.ptr.h))
}

func (s *Surface) At(x, y int) (c color.Color) {
	if image.Pt(x, y).In(s.Bounds()) {
		c = s.GetColor(s.Pixels32()[x+y*s.Pitch32()])
	}
	return
}

func (s *Surface) Set(x, y int, c color.Color) {
	if image.Pt(x, y).In(s.Bounds()) {
		s.Pixels32()[x+y*s.Pitch32()] = s.MapColor(c)
	}
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

func (s *Surface) Blit(x, y int, target *Surface) {
	targetRect := convertRect(image.Rect(x, y, x, y))
	C.SDL_BlitSurface(s.ptr, nil, target.ptr, &targetRect)
}

func (s *Surface) SetColorKey(c color.Color) {
	C.SDL_SetColorKey(s.ptr, C.SDL_SRCCOLORKEY, C.Uint32(s.MapColor(c)))
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
	C.SDL_FillRect(C.SDL_GetVideoSurface(), &sdlRect, C.Uint32(Video().MapColor(color)))
}

func Clear(color color.Color) {
	C.SDL_FillRect(C.SDL_GetVideoSurface(), nil, C.Uint32(Video().MapColor(color)))
}

func NewSurface(w, h int) (s *Surface) {
	video := C.SDL_GetVideoSurface()
	ptr := C.SDL_CreateRGBSurface(
		0, C.int(w), C.int(h), C.int(video.format.BitsPerPixel),
		video.format.Rmask,
		video.format.Gmask,
		video.format.Bmask,
		video.format.Amask)
	s = &Surface{ptr}
	runtime.SetFinalizer(s, func(s *Surface) { C.SDL_FreeSurface(s.ptr) })
	return
}

func ToSurface(img image.Image) (s *Surface) {
	s = NewSurface(img.Bounds().Dx(), img.Bounds().Dy())
	//s.Set(0, 0, color.RGBA{0xff, 0xff, 0xff, 0xff})
	draw.Draw(s, s.Bounds(), img, img.Bounds().Min, draw.Over)
	return
}
