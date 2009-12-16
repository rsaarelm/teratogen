package gfx

import (
	"image"
	"image/png"
	"once"
	"strings"
)

type DrawImage interface {
	image.Image
	Set(x, y int, c image.Color)
}

type Mask [][]byte

func (self Mask) Width() int { return len(self) }

func (self Mask) Height() int { return len(self[0]) }

// Convert the pixels beneath mask with return values of filter given the
// original pixel and the mask value.
func BlitMask(
	img DrawImage,
	mask Mask,
	filter func(maskVal byte, srcVal image.Color) (dstVal image.Color),
	ox, oy int) {
	for x, ex := 0, mask.Width(); x < ex; x++ {
		for y, ey := 0, mask.Height(); y < ey; y++ {
			xp, yp := x + ox, y + oy
			img.Set(xp, yp, filter(mask[x][y], img.At(xp, yp)))
		}
	}
}

// Use filter on surface pixels to turn the surface into mask.
func MakeMask(
	img image.Image,
	filter func(src image.Color) byte) (mask Mask) {
	mask = make(Mask, img.Width())
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
	mask Mask,
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

const errorImageData =
"\x89\x50\x4e\x47\x0d\x0a\x1a\x0a\x00\x00\x00\x0d\x49\x48\x44\x52"+
"\x00\x00\x00\x0c\x00\x00\x00\x0e\x08\x06\x00\x00\x00\x1b\xbd\xfd"+
"\xec\x00\x00\x00\x01\x73\x52\x47\x42\x00\xae\xce\x1c\xe9\x00\x00"+
"\x00\x06\x62\x4b\x47\x44\x00\xff\x00\xff\x00\xff\xa0\xbd\xa7\x93"+
"\x00\x00\x00\x09\x70\x48\x59\x73\x00\x00\x0b\x13\x00\x00\x0b\x13"+
"\x01\x00\x9a\x9c\x18\x00\x00\x00\x07\x74\x49\x4d\x45\x07\xd9\x0c"+
"\x0f\x14\x06\x3b\x98\x08\x5f\x1e\x00\x00\x00\x86\x49\x44\x41\x54"+
"\x28\xcf\x85\x52\xc1\x11\x04\x21\x08\x63\x6f\xb6\x0a\xeb\xa0\x0f"+
"\x6b\x0d\x75\xd0\x06\xb4\xc1\x3e\x6e\x74\x0e\x8f\xd5\x7c\x34\x43"+
"\x12\x19\xe1\x02\x10\x74\x40\xef\xfd\x9a\x04\x40\xbc\x01\x40\x98"+
"\x59\xfc\x86\x7e\x4e\xe9\xad\x35\x62\x66\x1a\xa6\xa3\x61\x35\xdd"+
"\x27\xb1\x88\x24\xbe\x35\x30\x73\xe2\xaa\xfa\x6d\x49\x44\xc8\xdd"+
"\x89\x88\xc8\xdd\x49\x44\x66\xb2\xaa\xce\xb6\x12\xaa\x1f\x19\xf7"+
"\xaa\x96\x04\x15\x52\x6d\x15\x6e\x87\xb9\x3e\x57\xf1\x71\x02\x88"+
"\x39\xe9\x21\x32\xb3\x3f\x9e\x42\x76\xab\x51\xad\xca\x5d\x0d\x67"+
"\x87\x07\x9d\xad\xdd\xe8\x00\x59\xe2\x1d\x00\x00\x00\x00\x49\x45"+
"\x4e\x44\xae\x42\x60\x82"

var errorImage image.Image

func loadError() {
// Use this image to indicate failed loading of an actual image.
	errorImage, _ = png.Decode(strings.NewReader(errorImageData))
}

func ErrorImage() image.Image {
	once.Do(loadError)
	return errorImage
}