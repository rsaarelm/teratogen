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
