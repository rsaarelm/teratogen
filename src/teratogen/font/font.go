/* font.go

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

// Package font reads TrueType font files and renders them into bitmaps that
// can be used in games. It uses the TrueType rasterizer v0.6c from Sean
// Barrett's STB library: http://nothings.org/stb/stb_truetype.h
package font

/*
// stb_truetype has a couple unused variables. Make cgo not worry about them.
#cgo CFLAGS: -Wno-error=unused-but-set-variable
#cgo LDFLAGS: -lm

// Include a whole C library as source code. Because cgo is magic.
#define STB_TRUETYPE_IMPLEMENTATION
#include "stb_truetype.h"
*/
import "C"

import (
	"errors"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"reflect"
	"teratogen/gfx"
	"unsafe"
)

type glyph struct {
	x0, y0, x1, y1       uint16
	xOff, yOff, xAdvance float32
}

type Font struct {
	pixels              []byte
	pixelsW, pixelsH    int
	startChar, numChars int

	glyphHeight float64
	glyphs      []glyph
}

// New creates a new bitmap font sheet with the desired characters rendered in
// the desired size from the given TTF font data in the byte buffer.
func New(
	r io.Reader, glyphHeight float64,
	startChar, numChars int) (result *Font, err error) {

	ttfBuffer, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	result = new(Font)

	// Estimate a sufficiently big power of two dimension for the font sheet.
	dim := 1
	for dim*dim < numChars*int(glyphHeight*glyphHeight) {
		dim *= 2
	}

	result.pixels = make([]byte, dim*dim)
	result.pixelsW = dim
	result.pixelsH = dim
	result.startChar = startChar
	result.numChars = numChars
	result.glyphHeight = glyphHeight
	result.glyphs = make([]glyph, numChars)

	inputPtr := ((*reflect.SliceHeader)(unsafe.Pointer(&ttfBuffer))).Data
	pixelsPtr := ((*reflect.SliceHeader)(unsafe.Pointer(&result.pixels))).Data
	glyphsPtr := ((*reflect.SliceHeader)(unsafe.Pointer(&result.glyphs))).Data

	ret := C.stbtt_BakeFontBitmap(
		(*C.uchar)(unsafe.Pointer(inputPtr)), 0, C.float(glyphHeight),
		(*C.uchar)(unsafe.Pointer(pixelsPtr)), C.int(dim), C.int(dim),
		C.int(startChar), C.int(numChars),
		(*C.stbtt_bakedchar)(unsafe.Pointer(glyphsPtr)))

	if ret <= 0 {
		err = errors.New("Couldn't create font sheet")
	}
	return
}

// Pixels returns the byte array of a font sheet's 8-bit pixel data.
func (s *Font) Pixels() []byte {
	return s.pixels
}

// Pitch returns the size in bytes of a font sheets horizontal scanline.
func (s *Font) Pitch() int {
	return s.pixelsW
}

func (s *Font) valid(ch rune) bool {
	return int(ch) >= s.startChar && int(ch) < s.startChar+s.numChars
}

func (s *Font) advance(ch rune) (result float64) {
	if s.valid(ch) {
		result = float64(s.glyphs[int(ch)-s.startChar].xAdvance)
	}
	return
}

func (s *Font) StringWidth(str string) (width float64) {
	for _, ch := range str {
		width += s.advance(ch)
	}
	return
}

func (s *Font) render32BitChar(
	ch rune, color uint32, x, y int, target gfx.Surface32Bit) {
	if !s.valid(ch) {
		return
	}

	g := s.glyphs[int(ch)-s.startChar]

	x += int(g.xOff)
	y += int(g.yOff)

	tPix := target.Pixels32()

	tRect := target.Bounds()

	for gy := 0; gy <= int(g.y1-g.y0); gy++ {
		tPos := x + (y+gy)*target.Pitch32()
		gPos := (gy+int(g.y0))*s.Pitch() + int(g.x0)

		for gx := 0; gx <= int(g.x1-g.x0); gx++ {
			if s.pixels[gPos+gx] >= 0x80 && image.Pt(x+gx, y+gy).In(tRect) {
				tPix[tPos+gx] = color
			}
		}
	}
}

// RenderTo32Bit renders a string using the bitmapped font to a non-antialised
// string to a 32-bit buffer with the given color. It returns the width of the
// string.
func (s *Font) RenderTo32Bit(
	str string, col color.Color, x, y int, target gfx.Surface32Bit) (xAdvance float64) {
	col32 := target.MapColor(col)
	for _, ch := range str {
		if !s.valid(ch) {
			continue
		}

		g := s.glyphs[int(ch)-s.startChar]

		s.render32BitChar(ch, col32, x+int(xAdvance), y, target)

		xAdvance += float64(g.xAdvance)
	}
	return
}
