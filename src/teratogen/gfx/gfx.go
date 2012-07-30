/* gfx.go

   Copyright (C) 2012 Risto Saarelma

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

// Package gfx provides miscellaneous graphics utilities.
package gfx

import (
	"fmt"
	"image"
	"image/color"
	"teratogen/sdl"
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

// ImageDrawable is a Drawable made from a SDL surface.
type ImageDrawable struct {
	Surface *sdl.Surface
	Rect    image.Rectangle
	Offset  image.Point
}

func (d ImageDrawable) Draw(offset image.Point) {
	offset = offset.Add(d.Offset)
	d.Surface.Blit(d.Rect, offset.X, offset.Y, sdl.Video())
}

func (d ImageDrawable) Bounds() image.Rectangle {
	return d.Rect.Add(d.Offset)
}

func (d ImageDrawable) String() string {
	return fmt.Sprintf("ImageDrawable %s", d.Bounds())
}

// LerpCol returns a linearly interpolated color between the two endpoint
// colors.
func LerpCol(c1, c2 color.Color, x float64) color.Color {
	r1, b1, g1, a1 := c1.RGBA()
	r2, b2, g2, a2 := c2.RGBA()
	return color.RGBA{
		uint8((float64(r1) + float64(r2-r1)*x) / 256),
		uint8((float64(g1) + float64(g2-g1)*x) / 256),
		uint8((float64(b1) + float64(b2-b1)*x) / 256),
		uint8((float64(a1) + float64(a2-a1)*x) / 256)}
}
