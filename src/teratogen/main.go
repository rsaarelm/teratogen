package main

import (
	"fmt"
	"teratogen/sdl"
)

func main() {
	sdl.Open(800, 600)
	defer sdl.Close()
	key := <-sdl.KeyboardChan()
	fmt.Printf("%s\n", sdl.TranslateScancode(key.Code))
}
