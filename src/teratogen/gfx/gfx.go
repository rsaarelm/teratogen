// gfx.go
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

// Package gfx provides miscellaneous graphics utilities.
package gfx

import (
	"fmt"
	"image"
	"image/color"
	"teratogen/num"
	"teratogen/sdl"
	"unsafe"
)

type Surface32Bit interface {
	Pixels32() []uint32
	Pitch32() int
	Bounds() image.Rectangle

	// MapColor converts color data into the internal format of the surface.
	MapColor(c color.Color) uint32

	// GetColor converts an internal color representation into a Color struct.
	GetColor(c32 uint32) color.Color
}

// Scaled returns a SDL surface where the graphics have been multiplied by an
// even multiple of dimensions. Useful for doubling or tripling pixel
// dimensions of small pixel art.
func Scaled(orig *sdl.Surface, scale image.Point) (result *sdl.Surface) {
	if scale.X < 1 || scale.Y < 1 {
		panic("Bad scale dimensions")
	}

	if scale == image.Pt(1, 1) {
		return orig
	}

	result = sdl.NewSurface(orig.Bounds().Dx()*scale.X, orig.Bounds().Dy()*scale.Y)

	oPix := orig.Pixels32()
	rPix := result.Pixels32()

	for oy := 0; oy < orig.Bounds().Dy(); oy++ {
		for ox := 0; ox < orig.Bounds().Dx(); ox++ {
			for ry := oy * scale.Y; ry < oy*scale.Y+scale.Y; ry++ {
				for rx := ox * scale.X; rx < ox*scale.X+scale.X; rx++ {
					rPix[rx+ry*result.Pitch32()] = oPix[ox+oy*orig.Pitch32()]
				}
			}
		}
	}
	return
}

func BlitX2(src, dest Surface32Bit) {
	srcPix := src.Pixels32()
	srcPitch := src.Pitch32()
	destPix := dest.Pixels32()
	destPitch := dest.Pitch32()

	w := src.Bounds().Dx()
	for y, ey := 0, src.Bounds().Dy(); y < ey; y++ {
		hline2X(srcPix[y*srcPitch:], destPix[y*2*destPitch:], w)
		hline2X(srcPix[y*srcPitch:], destPix[(y*2+1)*destPitch:], w)
	}
}

func BlitX3(src, dest Surface32Bit) {
	srcPix := src.Pixels32()
	srcPitch := src.Pitch32()
	destPix := dest.Pixels32()
	destPitch := dest.Pitch32()

	w := src.Bounds().Dx()
	for y, ey := 0, src.Bounds().Dy(); y < ey; y++ {
		hline3X(srcPix[y*srcPitch:], destPix[y*3*destPitch:], w)
		hline3X(srcPix[y*srcPitch:], destPix[(y*3+1)*destPitch:], w)
		hline3X(srcPix[y*srcPitch:], destPix[(y*3+2)*destPitch:], w)
	}
}

func GradientRect(s *sdl.Surface, rect image.Rectangle, topCol, bottomCol color.Color) {
	dy := rect.Dy()
	for y := 0; y < dy; y++ {
		s.FillRect(image.Rect(rect.Min.X, rect.Min.Y+y, rect.Max.X, rect.Min.Y+y+1),
			LerpCol(
				topCol,
				bottomCol,
				float64(y)/float64(dy)))
	}
}

// ImageDrawable is a Drawable made from a SDL surface.
type ImageDrawable struct {
	Surface *sdl.Surface
	Rect    image.Rectangle
	Offset  image.Point
}

func (d ImageDrawable) Draw(offset image.Point) {
	offset = offset.Add(d.Offset)
	d.Surface.Blit(d.Rect, offset.X, offset.Y, sdl.Frame())
}

func (d ImageDrawable) Bounds() image.Rectangle {
	return d.Rect.Add(d.Offset)
}

func (d ImageDrawable) String() string {
	return fmt.Sprintf("ImageDrawable %s", d.Bounds())
}

func hline2X(src, dest []uint32, n int) {
	srcPtr, destPtr := uintptr(unsafe.Pointer(&src[0])), uintptr(unsafe.Pointer(&dest[0]))
	for n > 0 {
		*(*uint32)(unsafe.Pointer(destPtr)) = *(*uint32)(unsafe.Pointer(srcPtr))
		destPtr += 4
		*(*uint32)(unsafe.Pointer(destPtr)) = *(*uint32)(unsafe.Pointer(srcPtr))
		destPtr += 4
		srcPtr += 4
		n--
	}
}

func hline3X(src, dest []uint32, n int) {
	srcPtr, destPtr := uintptr(unsafe.Pointer(&src[0])), uintptr(unsafe.Pointer(&dest[0]))
	for n > 0 {
		*(*uint32)(unsafe.Pointer(destPtr)) = *(*uint32)(unsafe.Pointer(srcPtr))
		destPtr += 4
		*(*uint32)(unsafe.Pointer(destPtr)) = *(*uint32)(unsafe.Pointer(srcPtr))
		destPtr += 4
		*(*uint32)(unsafe.Pointer(destPtr)) = *(*uint32)(unsafe.Pointer(srcPtr))
		destPtr += 4
		srcPtr += 4
		n--
	}
}

type ImageSpec struct {
	File   string
	Bounds image.Rectangle
	Offset image.Point
}

func SubImage(file string, bounds image.Rectangle) ImageSpec {
	return ImageSpec{file, bounds, image.Pt(0, 0)}
}

func OffsetSubImage(file string, bounds image.Rectangle, offset image.Point) ImageSpec {
	return ImageSpec{file, bounds, offset}
}

// Context is a interface for Spritable objects to get UI level resources to
// turn their abstract representation data into actual drawable assets such as
// bitmap surface handles.
type Context interface {
	// GetDrawable converts an ImageSpec into a Drawable object, probably by
	// fetching it from some sort of cache.
	GetDrawable(spec ImageSpec) Drawable
}

type Drawable interface {
	Draw(offset image.Point)
}

func Line(s *sdl.Surface, p1, p2 image.Point, col color.Color) {
	num.BresenhamLine(func(p image.Point) { s.Set(p.X, p.Y, col) }, p1, p2)
}
