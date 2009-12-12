package main

import (
	"hyades/sdl"
	"image"
	"time"
)

func main() {
	sdl.InitSdl(640, 480, "Hello SDL", false)

	sprite := sdl.Make32BitSurface(0, 32, 32)

	for x := 0; x < 32; x++ {
		for y := 0; y < 32; y++ {
			sprite.Set(x, y, image.RGBAColor{byte(x*8), byte(y*8), 255, 255})
		}
	}
	sprite.Blit(sdl.GetVideoSurface(), 128, 32)
	sdl.Flip()
	time.Sleep(2e9)
	sdl.ExitSdl()
}
