// pal_test.go
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

package gfx

import (
	"image/color"
	"teratogen/sdl"
	"testing"
)

func unpaletted() (result *sdl.Surface) {
	result = sdl.NewSurface(8, 8)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			result.Set(x, y, color.RGBA{byte(x*16 + y*16), 0, 0, 255})
		}
	}
	return result
}

var pal = sdl.MakePalette([]color.Color{
	color.RGBA{0, 0x00, 0, 0xff},
	color.RGBA{0, 0x10, 0, 0xff},
	color.RGBA{0, 0x20, 0, 0xff},
	color.RGBA{0, 0x30, 0, 0xff},
	color.RGBA{0, 0x40, 0, 0xff},
	color.RGBA{0, 0x50, 0, 0xff},
	color.RGBA{0, 0x60, 0, 0xff},
	color.RGBA{0, 0x70, 0, 0xff},
	color.RGBA{0, 0x80, 0, 0xff},
	color.RGBA{0, 0x90, 0, 0xff},
	color.RGBA{0, 0xa0, 0, 0xff},
	color.RGBA{0, 0xb0, 0, 0xff},
	color.RGBA{0, 0xc0, 0, 0xff},
	color.RGBA{0, 0xd0, 0, 0xff},
	color.RGBA{0, 0xe0, 0, 0xff},
	color.RGBA{0, 0xf0, 0, 0xff}})

func paletted() (result *sdl.Surface) {
	result = sdl.NewPaletteSurface(8, 8)
	result.SetColors(pal)
	for i := 0; i < 64; i++ {
		result.Pixels8()[i] = byte(i % 16)
	}
	return result
}

func BenchmarkPalettedBlit(b *testing.B) {
	b.StopTimer()
	sdl.Run(320, 240)
	defer sdl.Stop()
	surf := paletted()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < 30; y++ {
			for x := 0; x < 40; x++ {
				surf.Blit(surf.Bounds(), x*8, y*8, sdl.Frame())
			}
		}
		sdl.Flip()
	}
	b.StopTimer()
}

func BenchmarkRepalettingBlit(b *testing.B) {
	b.StopTimer()
	sdl.Run(320, 240)
	defer sdl.Stop()
	surf := paletted()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < 30; y++ {
			for x := 0; x < 40; x++ {
				surf.Blit(surf.Bounds(), x*8, y*8, sdl.Frame())
				surf.SetColors(pal)
			}
		}
		sdl.Flip()
	}
	b.StopTimer()
}

func BenchmarkTruecolorBlit(b *testing.B) {
	b.StopTimer()
	sdl.Run(320, 240)
	defer sdl.Stop()
	surf := unpaletted()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < 30; y++ {
			for x := 0; x < 40; x++ {
				surf.Blit(surf.Bounds(), x*8, y*8, sdl.Frame())
			}
		}
		sdl.Flip()
	}
	b.StopTimer()
}
