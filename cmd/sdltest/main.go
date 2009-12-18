package main

import (
	"fmt"
	"hyades/event"
	"hyades/sdl"
	"hyades/sfx"
	"image"
	"strings"
)

// A png sprite by Oddball for the Assemblee contest at tigsource
// (http://forums.tigsource.com/index.php?topic=8834.0)
const Elf_png =
"\x89\x50\x4e\x47\x0d\x0a\x1a\x0a\x00\x00\x00\x0d\x49\x48\x44\x52" +
"\x00\x00\x00\x08\x00\x00\x00\x08\x08\x06\x00\x00\x00\xc4\x0f\xbe" +
"\x8b\x00\x00\x00\x01\x73\x52\x47\x42\x00\xae\xce\x1c\xe9\x00\x00" +
"\x00\x79\x49\x44\x41\x54\x18\xd3\x63\x5c\x7c\x4e\xfd\x3f\x03\x1a" +
"\x88\x35\xba\xc9\xb8\x24\x45\xe9\x7f\xcc\x9c\x7b\x8c\x8c\x0c\x0c" +
"\x0c\x0c\xf8\x14\xb1\xc0\x04\x42\x44\x5c\x18\x38\xe5\xa6\x32\x7c" +
"\x7f\x94\xcd\xc0\x29\x37\x15\xae\x90\x09\x26\xc9\xc0\xc0\xc0\xf0" +
"\xfd\x51\x36\x0a\xcd\xc0\xc0\xc0\xc0\x88\xcd\x78\x64\x53\x59\x42" +
"\x44\x5c\x18\xd6\x36\x6d\x87\x0b\xfe\xb4\xac\x66\x60\x3f\xde\xca" +
"\x10\x33\xe7\x1e\x23\x63\xca\xef\xff\x04\x1d\xc9\xc0\xc0\xc0\xc0" +
"\x00\x67\x40\xc1\x92\x14\xa5\xff\x30\x31\x00\x02\xa6\x31\x83\x52" +
"\x2e\xa8\xf4\x00\x00\x00\x00\x49\x45\x4e\x44\xae\x42\x60\x82"

func main() {
	sdl.Init(640, 480, "Hello SDL", false)

	sprite, err := sdl.MakePngSurface(strings.NewReader(Elf_png))

	if err != nil {
		panic("Image loading error" + err.String())
	}

	sprite2 := doubleSprite(sprite)
	sprite.FreeSurface()
	sprite2.Convert(sdl.GetVideoSurface())

	sfxTest()

	Outer: for {
		sdl.GetVideoSurface().FillRect(sdl.Rect(0, 0, 320, 240), image.RGBAColor{0, 0, 96, 255})
		sprite2.Blit(sdl.GetVideoSurface(), 128, 32)
		sdl.Flip()
		switch evt := sdl.PollEvent().(type) {
		case *event.KeyDown:
			fmt.Printf("%T: %+v\n", evt, evt)
			if evt.KeySym == event.K_Q { break Outer }
		case *event.Quit:
			break Outer
		default:
			if evt != nil {
				fmt.Printf("%T: %+v\n", evt, evt)
			}
		}
	}

	sdl.Exit()
}

func doubleSprite(src *sdl.Surface) (dst *sdl.Surface) {
	dst = sdl.Make32BitSurface(0, src.Width() * 2, src.Height() * 2)
	for x := 0; x < dst.Width(); x++ {
		for y := 0; y < dst.Height(); y++ {
			dst.Set(x, y, src.At(x / 2, y / 2))
		}
	}
	return
}

func sfxTest() {
	wave := sfx.MakeMono8Wav(
		sfx.ADSRFilter(0.2, 0.1, 0.8, 0.5, 0.3,
		sfx.MakeWave(1000.0, sfx.Jump(0.4, 400.0, sfx.Sine))),
		sdl.AudioRateHz(),
		1.0)
	sfx, err := sdl.LoadWav(wave)

	if err != nil {
		panic("Wav loading error: "+err.String())
	}

	sfx.Play(0)
}
