package main

import (
	"fmt"
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

	font.RenderTo32Bit("Hello, world!", 0xffffffff, 32, 32, sdl.Video())
	sdl.Flip()

	key := <-sdl.KeyboardChan()
	fmt.Printf("%s\n", key.FixedSym())
}
