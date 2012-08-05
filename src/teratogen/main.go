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
	sdl.Open(960, 720)
	defer sdl.Close()

	sdl.EnableKeyRepeat(sdl.DefaultRepeatDelay, sdl.DefaultRepeatInterval)

	fs, err := initArchive()
	if err != nil {
		panic(err)
	}
	ch := cache.New(fs)

	sdl.SetFrame(sdl.NewSurface(320, 240))

	w := world.New()
	w.TestMap(manifold.Location{0, 0, 1})

	disp := display.New(ch, w)

	gfx.GradientRect(sdl.Frame(), image.Rect(0, 0, 320, 160), gfx.Green, gfx.ScaleCol(gfx.Green, 0.2))
	disp.DrawWorld(image.Rect(4, 4, 316, 156))
	disp.DrawMsg(image.Rect(0, 160, 160, 240))

	gfx.BlitX3(sdl.Frame(), sdl.Video())
	sdl.Flip()

	for {
		switch e := (<-sdl.Events).(type) {
		case sdl.KeyEvent:
			if e.Sym == sdl.K_ESCAPE {
				return
			}
			if e.KeyDown {
				switch e.Sym {
				case sdl.K_LEFT:
					disp.Move(image.Pt(-1, 0))
				case sdl.K_RIGHT:
					disp.Move(image.Pt(1, 0))
				case sdl.K_UP:
					disp.Move(image.Pt(0, -1))
				case sdl.K_DOWN:
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
