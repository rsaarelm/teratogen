package gfx

import (
	"image"
)

type DrawImage interface {
	image.Image
	Set(x, y int, c image.Color)
}

// Convert the pixels beneath mask with return values of filter given the
// original pixel and the mask value.
func BlitMask(
	img DrawImage,
	mask [][]byte,
	filter func(maskVal byte, srcVal image.Color) (dstVal image.Color),
	ox, oy int) {
	for x, ex := 0, len(mask); x < ex; x++ {
		for y, ey := 0, len(mask[x]); y < ey; y++ {
			xp, yp := x + ox, y + oy
			img.Set(xp, yp, filter(mask[x][y], img.At(xp, yp)))
		}
	}
}

// Use filter on surface pixels to turn the surface into mask.
func MakeMask(
	img image.Image,
	filter func(src image.Color) byte) (mask [][]byte) {
	mask = make([][]byte, img.Width())
	for x := 0; x < img.Width(); x++ {
		mask[x] = make([]byte, img.Height())
		for y := 0; y < img.Height(); y++ {
			mask[x][y] = filter(img.At(x, y))
		}
	}
	return
}

func BlitColorMask(
	img DrawImage,
	mask [][]byte,
	col image.Color,
	ox, oy int) {
	BlitMask(img,
		mask,
		func (maskVal byte, srcVal image.Color) image.Color {
			if maskVal > 127 { return col }
			return srcVal
		},
		ox, oy)
}
