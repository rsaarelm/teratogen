/* main.go

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

package main

import (
	"fmt"
	"image"
	"os"
	"teratogen/archive"
	"teratogen/cache"
	"teratogen/font"
	"teratogen/gfx"
	"teratogen/manifold"
	"teratogen/sdl"
	"teratogen/world"
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
	sdl.Open(960, 720)
	defer sdl.Close()

	fs, err := initArchive()
	if err != nil {
		panic(err)
	}
	ch := cache.New(fs)

	sdl.SetFrame(sdl.NewSurface(320, 240))

	f, err := ch.GetFont(cache.FontSpec{"assets/04round_bold.ttf", 8.0, 32, 96})
	if err != nil {
		panic(err)
	}

	w := world.New()
	w.TestMap(manifold.Location{0, 0, 1})

	sprite, _ := ch.GetImage(cache.ImageSpec{"assets/chars.png", image.Rect(0, 8, 8, 16)})

	sprite.Draw(image.Pt(16, 16))

	gfx.GradientRect(sdl.Frame(), image.Rect(32, 32, 110, 44), gfx.Gold, gfx.ScaleCol(gfx.Gold, 0.5))

	cur := &font.Cursor{f, sdl.Frame(), image.Pt(36, 40), font.Emboss, gfx.Yellow, gfx.ScaleCol(gfx.Gold, 0.2)}

	fmt.Fprintf(cur, "Hello, world!")

	gfx.BlitX3(sdl.Frame(), sdl.Video())
	sdl.Flip()

	for {
		switch e := (<-sdl.Events).(type) {
		case sdl.KeyEvent:
			fmt.Printf("%s\n", e)
			if e.Sym == sdl.K_ESCAPE {
				return
			}
		case sdl.QuitEvent:
			return
		}
	}
}
