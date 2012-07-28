package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"teratogen/font"
	"teratogen/sdl"
)

func main() {
	sdl.Open(800, 600)
	defer sdl.Close()

	fontBuf, err := ioutil.ReadFile("assets/04round_bold.ttf")
	if err != nil {
		panic(err)
	}

	font, err := font.New(fontBuf, 16.0, 32, 96)
	if err != nil {
		panic(err)
	}

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
