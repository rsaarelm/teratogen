package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"teratogen/archive"
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
	sdl.Open(800, 600)
	defer sdl.Close()

	fs, err := initArchive()
	if err != nil {
		panic(err)
	}

	font, err := archive.LoadFont(fs, "assets/04round_bold.ttf", 16.0, 32, 96)
	if err != nil {
		panic(err)
	}

	pic, err := archive.LoadPng(fs, "assets/chars.png")

	sdlPic := gfx.Scaled(sdl.ToSurface(pic), 2, 2)
	sdlPic.SetColorKey(color.RGBA{0x00, 0xff, 0xff, 0xff})
	sdlPic.Blit(0, 0, sdl.Video())

	draw.Draw(sdl.Video(), image.Rect(300, 0, 800, 600), pic, image.Pt(0, 0), draw.Over)

	font.RenderTo32Bit(
		"Hello, world!",
		sdl.Video().MapColor(color.RGBA{128, 255, 128, 255}),
		32, 32, sdl.Video())
	sdl.Flip()

	for {
		key := <-sdl.KeyboardChan()
		fmt.Printf("%s\n", key)
		if key.Sym == sdl.K_ESCAPE {
			break
		}
	}
}
