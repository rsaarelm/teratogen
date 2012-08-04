// cache.go
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

package cache

import (
	"image"
	"image/color"
	"teratogen/archive"
	"teratogen/font"
	"teratogen/gfx"
	"teratogen/sdl"
)

type Cache struct {
	fs       archive.Device
	surfaces map[surfaceSpec]*sdl.Surface
	fonts    map[FontSpec]*font.Font
}

func New(fs archive.Device) (result *Cache) {
	result = new(Cache)
	result.fs = fs
	result.surfaces = make(map[surfaceSpec]*sdl.Surface)
	result.fonts = make(map[FontSpec]*font.Font)
	return
}

func (c *Cache) GetImage(spec ImageSpec) (result gfx.ImageDrawable, err error) {
	surface, err := c.getSurface(surfaceSpec{spec.File})
	if err != nil {
		return
	}
	return gfx.ImageDrawable{surface, spec.Bounds, image.Pt(0, 0)}, nil
}

func (c *Cache) GetFont(spec FontSpec) (result *font.Font, err error) {
	result, ok := c.fonts[spec]
	if !ok {
		result, err = archive.LoadFont(c.fs, spec.File, spec.Height, spec.BeginChar, spec.NumChars)
		if err != nil {
			return
		}
		c.fonts[spec] = result
	}
	return
}

func (c *Cache) getSurface(spec surfaceSpec) (result *sdl.Surface, err error) {
	result, ok := c.surfaces[spec]
	if !ok {
		var png image.Image
		png, err = archive.LoadPng(c.fs, spec.File)
		if err != nil {
			return
		}
		result = sdl.ToSurface(png)

		// XXX: Hardcoding the same colorkey for all images.
		result.SetColorKey(color.RGBA{0x00, 0xff, 0xff, 0xff})

		c.surfaces[spec] = result
	}
	return
}

type surfaceSpec struct {
	File string
}

type ImageSpec struct {
	File   string
	Bounds image.Rectangle
}

type FontSpec struct {
	File      string
	Height    float64
	BeginChar int
	NumChars  int
}
