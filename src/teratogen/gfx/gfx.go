package gfx

import (
	"image"
	"image/color"
	"teratogen/sdl"
)

type Surface32Bit interface {
	Pixels32() []uint32
	Pitch32() int
	Bounds() image.Rectangle

	// MapColor converts color data into the internal format of the surface.
	MapColor(c color.Color) uint32

	// GetColor converts an internal color representation into a Color struct.
	GetColor(c32 uint32) color.Color
}

func Scaled(orig *sdl.Surface, sx, sy int) (result *sdl.Surface) {
	if sx < 1 || sy < 1 {
		panic("Bad scale dimensions")
	}

	result = sdl.NewSurface(orig.Bounds().Dx()*sx, orig.Bounds().Dy()*sy)

	oPix := orig.Pixels32()
	rPix := result.Pixels32()

	for oy := 0; oy < orig.Bounds().Dy(); oy++ {
		for ox := 0; ox < orig.Bounds().Dx(); ox++ {
			for ry := oy * sy; ry < oy*sy+sy; ry++ {
				for rx := ox * sx; rx < ox*sx+sx; rx++ {
					rPix[rx+ry*result.Pitch32()] = oPix[ox+oy*orig.Pitch32()]
				}
			}
		}
	}
	return
}
