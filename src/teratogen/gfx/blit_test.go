/* blit_test.go

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

package gfx

import (
	"teratogen/sdl"
	"testing"
)

// Benchmark various zooming and non-zooming screen blits.

func BenchmarkNoZoomBlit(b *testing.B) {
	b.StopTimer()
	sdl.Open(320, 240)
	defer sdl.Close()
	sdl.SetFrame(sdl.NewSurface(320, 240))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sdl.Frame().Blit(sdl.Frame().Bounds(), 0, 0, sdl.Video())
		sdl.Flip()
	}
	b.StopTimer()
}

func Benchmark2XZoomBlit(b *testing.B) {
	b.StopTimer()
	sdl.Open(640, 480)
	defer sdl.Close()
	sdl.SetFrame(sdl.NewSurface(320, 240))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		BlitX2(sdl.Frame(), sdl.Video())
		sdl.Flip()
	}
	b.StopTimer()
}

func BenchmarkNoZoom2XBlit(b *testing.B) {
	b.StopTimer()
	sdl.Open(640, 480)
	defer sdl.Close()
	sdl.SetFrame(sdl.NewSurface(640, 480))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sdl.Frame().Blit(sdl.Frame().Bounds(), 0, 0, sdl.Video())
		sdl.Flip()
	}
	b.StopTimer()
}

func Benchmark3XZoomBlit(b *testing.B) {
	b.StopTimer()
	sdl.Open(960, 720)
	defer sdl.Close()
	sdl.SetFrame(sdl.NewSurface(320, 240))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		BlitX3(sdl.Frame(), sdl.Video())
		sdl.Flip()
	}
	b.StopTimer()
}

func BenchmarkNoZoom3XBlit(b *testing.B) {
	b.StopTimer()
	sdl.Open(960, 720)
	defer sdl.Close()
	sdl.SetFrame(sdl.NewSurface(960, 720))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sdl.Frame().Blit(sdl.Frame().Bounds(), 0, 0, sdl.Video())
		sdl.Flip()
	}
	b.StopTimer()
}
