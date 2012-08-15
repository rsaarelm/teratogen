// main.go
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

package main

import (
	"image"
	"math/rand"
	"os"
	"teratogen/archive"
	"teratogen/cache"
	"teratogen/display"
	"teratogen/gfx"
	"teratogen/manifold"
	"teratogen/mob"
	"teratogen/sdl"
	"teratogen/world"
	"time"
)

// Set up a file archive that first looks for files in the local physical
// filesystem path, then in a zip file contained in the local binary.
func initArchive() (fs archive.Device, err error) {
	var devices = make([]archive.Device, 0)

	fd, err := archive.FsDevice(".")
	if err != nil {
		// If the file system path won't work, things are bad.
		return
	}
	devices = append(devices, fd)

	zd, zerr := archive.FileZipDevice(os.Args[0])
	// If the self exe isn't a zip, just don't add the device. Things still
	// work if the assets can be found in the filesystem.
	if zerr == nil {
		devices = append(devices, zd)
	}

	return archive.New(devices...), nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	sdl.Run(960, 720)
	defer sdl.Stop()

	fs, err := initArchive()
	if err != nil {
		panic(err)
	}
	ch := cache.New(fs)

	sdl.SetFrame(sdl.NewSurface(320, 240))

	w := world.New()
	origin := manifold.Location{0, 0, 1}
	w.TestMap(origin)

	bounds := image.Rect(-16, -16, 16, 16)
	for i := 0; i < 32; i++ {
		pos := image.Pt(rand.Intn(bounds.Dx())+bounds.Min.X, rand.Intn(bounds.Dy())+bounds.Min.Y)
		m := mob.New(w, &mob.Spec{gfx.ImageSpec{"assets/chars.png", image.Rect(8, 0, 16, 8)}})
		loc := origin.Add(pos)
		if m.Fits(loc) {
			m.Place(loc)
		}
	}

	pc := mob.New(w, &mob.Spec{gfx.ImageSpec{"assets/chars.png", image.Rect(0, 8, 8, 16)}})

found:
	for {
		for i := 0; i < 64; i++ {
			pos := image.Pt(rand.Intn(bounds.Dx())+bounds.Min.X, rand.Intn(bounds.Dy())+bounds.Min.Y)
			loc := origin.Add(pos)
			if pc.Fits(loc) {
				pc.Place(loc)
				w.Player = pc
				break found
			}
		}
		panic("Can't place player")
	}

	disp := display.New(ch, w)

	gfx.GradientRect(sdl.Frame(), image.Rect(0, 0, 320, 160), gfx.Green, gfx.ScaleCol(gfx.Green, 0.2))
	disp.DrawWorld(image.Rect(2, 2, 318, 158))
	disp.DrawMsg(image.Rect(2, 162, 158, 238))

	gfx.BlitX3(sdl.Frame(), sdl.Video())
	sdl.Flip()

	for {
		switch e := (<-sdl.Events).(type) {
		case sdl.KeyEvent:
			if e.Sym == sdl.K_ESCAPE {
				return
			}
			if e.KeyDown {
				// Layout independent keys
				switch e.FixedSym() {
				case sdl.K_q:
					disp.Move(image.Pt(-1, 0))
				case sdl.K_w:
					disp.Move(image.Pt(-1, -1))
				case sdl.K_e:
					disp.Move(image.Pt(0, -1))
				case sdl.K_d:
					disp.Move(image.Pt(1, 0))
				case sdl.K_s:
					disp.Move(image.Pt(1, 1))
				case sdl.K_a:
					disp.Move(image.Pt(0, 1))
				}
			}
		case sdl.QuitEvent:
			return
		}

		disp.DrawWorld(image.Rect(4, 4, 316, 156))
		gfx.BlitX3(sdl.Frame(), sdl.Video())
		sdl.Flip()
	}
}
