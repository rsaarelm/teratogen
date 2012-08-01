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
	"image/color"
	"os"
	"teratogen/archive"
	"teratogen/cache"
	"teratogen/font"
	"teratogen/gfx"
	"teratogen/sdl"
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
	sdl.Open(640, 480)
	defer sdl.Close()

	fs, err := initArchive()
	if err != nil {
		panic(err)
	}
	ch := cache.New(fs)

	f, err := ch.GetFont(cache.FontSpec{"assets/04round_bold.ttf", 16.0, 32, 96})
	if err != nil {
		panic(err)
	}

	sprite, _ := ch.GetImage(cache.ImageSpec{"assets/chars.png", image.Rect(0, 16, 16, 32), image.Pt(2, 2)})

	sprite.Draw(image.Pt(32, 32))

	for y := 0; y < 24; y++ {
		sdl.Video().FillRect(image.Rect(64, 64+y, 220, 64+y+1),
			gfx.LerpCol(
				color.RGBA{0xff, 0x50, 0x50, 0xff},
				color.RGBA{0x60, 0x30, 0x30, 0xff},
				float64(y)/24))
	}

	cur := &font.Cursor{f, sdl.Video(), image.Pt(72, 80), font.Emboss, color.RGBA{0xff, 0xdd, 0xdd, 0xff}, color.RGBA{0x22, 0x00, 0x00, 0xff}, 2}

	fmt.Fprintf(cur, "Hello, world!")

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
