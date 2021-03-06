// video.go
//
// Copyright (C) 2012 Risto Saarelma
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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

// Pixels32 maps an array of 32-bit values to the pixels of a 32-bit surface.
func (s *Surface) Pixels32() (result []uint32) {
	size := int(s.ptr.pitch) * int(s.ptr.h) / 4

	result = []uint32{}
	header := (*reflect.SliceHeader)(unsafe.Pointer(&result))
	*header = reflect.SliceHeader{uintptr(s.ptr.pixels), size, size}
	return
}

func (s *Surface) Pixels8() (result []uint8) {
	size := int(s.ptr.pitch) * int(s.ptr.h)

	result = []uint8{}
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

func (s *Surface) Blit(bounds image.Rectangle, x, y int, target *Surface) {
	mutex.Lock()
	defer mutex.Unlock()

	targetRect := convertRect(image.Rect(x, y, x, y))
	sourceRect := convertRect(bounds)
	C.SDL_BlitSurface(s.ptr, &sourceRect, target.ptr, &targetRect)
}

func (s *Surface) SetColorKey(c color.Color) {
	C.SDL_SetColorKey(s.ptr, C.SDL_SRCCOLORKEY, C.Uint32(s.MapColor(c)))
}

var frame *Surface

// SetFrame sets the default frame buffer to draw to.
func SetFrame(surface *Surface) {
	frame = surface
}

func Frame() *Surface {
	if frame == nil {
		return Video()
	}
	return frame
}

// Video returns the surface for the base SDL window.
func Video() *Surface {
	return &Surface{C.SDL_GetVideoSurface()}
}

// Flip swaps screen buffers with a double-buffered display mode. Use it to
// make the changes you made to the screen become visible.
func Flip() {
	mutex.Lock()
	defer mutex.Unlock()

	C.SDL_Flip(C.SDL_GetVideoSurface())
}

func (s *Surface) FillRect(rect image.Rectangle, color color.Color) {
	mutex.Lock()
	defer mutex.Unlock()

	sdlRect := convertRect(rect)
	C.SDL_FillRect(s.ptr, &sdlRect, C.Uint32(s.MapColor(color)))
}

func (s *Surface) Clear(color color.Color) {
	mutex.Lock()
	defer mutex.Unlock()

	C.SDL_FillRect(s.ptr, nil, C.Uint32(s.MapColor(color)))
}

func (s *Surface) SetClipRect(rect image.Rectangle) {
	r := convertRect(rect)
	C.SDL_SetClipRect(s.ptr, &r)
}

func (s *Surface) ClearClipRect() {
	C.SDL_SetClipRect(s.ptr, nil)
}

func (s *Surface) IsPalettized() bool {
	return s.ptr.format.BitsPerPixel == 8
}

type Palette []struct {
	R, G, B byte
	Unused  byte
}

func MakePalette(colors []color.Color) (result Palette) {
	result = make(Palette, len(colors))
	for i, _ := range result {
		r, g, b, _ := colors[i].RGBA()
		result[i].R, result[i].G, result[i].B = byte(r>>8), byte(g>>8), byte(b>>8)
	}
	return
}

func (s *Surface) SetColors(pal Palette) {
	colPtr := (*C.SDL_Color)(unsafe.Pointer(&pal[0]))
	C.SDL_SetPalette(s.ptr, C.SDL_LOGPAL, colPtr, 0, C.int(len(pal)))
}

func NewSurface(w, h int) (s *Surface) {
	mutex.Lock()
	defer mutex.Unlock()

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

func NewPaletteSurface(w, h int) (s *Surface) {
	mutex.Lock()
	defer mutex.Unlock()

	ptr := C.SDL_CreateRGBSurface(0, C.int(w), C.int(h), 8, 0, 0, 0, 0)
	s = &Surface{ptr}
	runtime.SetFinalizer(s, func(s *Surface) { C.SDL_FreeSurface(s.ptr) })
	return
}
