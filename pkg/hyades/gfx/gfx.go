// Rendering platform agnostic graphics operations

package gfx

import (
	"exp/draw"
	"hyades/dbg"
	"hyades/geom"
	"image"
	"image/png"
	"hyades/num"
	"os"
	"strings"
	"sync"
)

// Graphics context interface, for drawing into things.
type Graphics interface {
	draw.Image
	Blit(img image.Image, x, y int)
	FillRect(rect image.Rectangle, color image.Color)

	// Set a clipping rectangle on the context. Nothing will be drawn outside
	// the clipping rectangle.
	SetClip(clipRect image.Rectangle)

	// Clear the clipping rectangle.
	ClearClip()
}

type Constructor func(width, height int) draw.Image

// An Image implementation from a function that maps ([0..1), [0..1)) to RGBA.
type procImage struct {
	colorF        func(x, y float64) image.Color
	width, height int
}

func ProceduralImage(colorF func(float64, float64) image.Color, width, height int) image.Image {
	dbg.Assert(width > 0 && height > 0, "Procedural Image must have nonzero dimensions.")
	return &procImage{colorF, width, height}
}

func (self *procImage) ColorModel() image.ColorModel {
	return image.RGBAColorModel
}

func (self *procImage) Bounds() image.Rectangle { return image.Rect(0, 0, self.width, self.height) }

func (self *procImage) At(x, y int) image.Color {
	return self.colorF(float64(x)/float64(self.width-1), float64(y)/float64(self.height-1))
}

// An image filter that returns the contents of a source image exactly as they are.
func IdFilter(src image.Image) (result func(float64, float64) image.Color) {
	return func(x, y float64) image.Color {
		return src.At(
			int(num.Round(x*float64(src.Bounds().Dx()-1))),
			int(num.Round(y*float64(src.Bounds().Dy()-1))))
	}
}

// IntScaleImage creates a copy of an image that has been scaled using
// nearest-neighbor with an integer multiplication factor. Useful for making
// graphics with large uniform pixels.
func IntScaleImage(src image.Image, xScale, yScale int) image.Image {
	result := DefaultConstructor(src.Bounds().Dx()*xScale, src.Bounds().Dy()*yScale)
	for x := 0; x < result.Bounds().Dx(); x++ {
		for y := 0; y < result.Bounds().Dy(); y++ {
			result.Set(x, y, src.At(x/xScale, y/yScale))
		}
	}
	return result
}

type Mask [][]byte

func (self Mask) Bounds() image.Rectangle { return image.Rect(0, 0, len(self), len(self[0])) }

// Convert the pixels beneath mask with return values of filter given the
// original pixel and the mask value.
func BlitMask(img draw.Image, mask Mask, filter func(maskVal byte, srcVal image.Color) (dstVal image.Color), ox, oy int) {
	for x, ex := 0, mask.Bounds().Dx(); x < ex; x++ {
		for y, ey := 0, mask.Bounds().Dy(); y < ey; y++ {
			xp, yp := x+ox, y+oy
			img.Set(xp, yp, filter(mask[x][y], img.At(xp, yp)))
		}
	}
}

// Use filter on surface pixels to turn the surface into mask.
func MakeMask(img image.Image, filter func(src image.Color) byte) (mask Mask) {
	mask = make(Mask, img.Bounds().Dy())
	for x := 0; x < img.Bounds().Dx(); x++ {
		mask[x] = make([]byte, img.Bounds().Dy())
		for y := 0; y < img.Bounds().Dy(); y++ {
			mask[x][y] = filter(img.At(x, y))
		}
	}
	return
}

func BlitColorMask(img draw.Image, mask Mask, col image.Color, ox, oy int) {
	BlitMask(img,
		mask,
		func(maskVal byte, srcVal image.Color) image.Color {
			if maskVal > 127 {
				return col
			}
			return srcVal
		},
		ox, oy)
}

func FilterImage(img draw.Image, filter func(image.Color) image.Color) {
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			img.Set(x, y, filter(img.At(x, y)))
		}
	}
}

func FilterTransparent(img draw.Image, transparencyColor image.Color) {
	FilterImage(
		img,
		func(c image.Color) image.Color {
			if ColorsEqual(c, transparencyColor) {
				return image.RGBAColor{0, 0, 0, 0}
			}
			return c
		})
}

// AlphaRemoveFn is a FilterImage function that turns the alpha channel opaque.
func OpaqueAlphaFn(col image.Color) image.Color {
	r, g, b, _ := RGBA8Bit(col)
	return image.RGBAColor{r, g, b, 255}
}

func Clip(src image.Image, cons Constructor, rect image.Rectangle) (result draw.Image) {
	result = cons(rect.Dx(), rect.Dy())
	draw.Draw(result, result.Bounds(), src, rect.Min)
	return
}

func MakeTiles(src image.Image, cons Constructor, tileW, tileH int) (result []draw.Image) {
	cols, rows := src.Bounds().Dx()/tileW, src.Bounds().Dy()/tileH
	result = make([]draw.Image, cols*rows)
	i := 0
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			result[i] = Clip(src, cons,
				image.Rect(x*tileW, y*tileH, (x+1)*tileW, (y+1)*tileH))
			i++
		}
	}
	return
}

func DefaultConstructor(width, height int) draw.Image {
	return image.NewRGBA(width, height)
}

func ColorsEqual(a, b image.Color) bool {
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

const errorImageData = "\x89\x50\x4e\x47\x0d\x0a\x1a\x0a\x00\x00\x00\x0d\x49\x48\x44\x52" +
	"\x00\x00\x00\x0c\x00\x00\x00\x0e\x08\x06\x00\x00\x00\x1b\xbd\xfd" +
	"\xec\x00\x00\x00\x01\x73\x52\x47\x42\x00\xae\xce\x1c\xe9\x00\x00" +
	"\x00\x06\x62\x4b\x47\x44\x00\xff\x00\xff\x00\xff\xa0\xbd\xa7\x93" +
	"\x00\x00\x00\x09\x70\x48\x59\x73\x00\x00\x0b\x13\x00\x00\x0b\x13" +
	"\x01\x00\x9a\x9c\x18\x00\x00\x00\x07\x74\x49\x4d\x45\x07\xd9\x0c" +
	"\x0f\x14\x06\x3b\x98\x08\x5f\x1e\x00\x00\x00\x86\x49\x44\x41\x54" +
	"\x28\xcf\x85\x52\xc1\x11\x04\x21\x08\x63\x6f\xb6\x0a\xeb\xa0\x0f" +
	"\x6b\x0d\x75\xd0\x06\xb4\xc1\x3e\x6e\x74\x0e\x8f\xd5\x7c\x34\x43" +
	"\x12\x19\xe1\x02\x10\x74\x40\xef\xfd\x9a\x04\x40\xbc\x01\x40\x98" +
	"\x59\xfc\x86\x7e\x4e\xe9\xad\x35\x62\x66\x1a\xa6\xa3\x61\x35\xdd" +
	"\x27\xb1\x88\x24\xbe\x35\x30\x73\xe2\xaa\xfa\x6d\x49\x44\xc8\xdd" +
	"\x89\x88\xc8\xdd\x49\x44\x66\xb2\xaa\xce\xb6\x12\xaa\x1f\x19\xf7" +
	"\xaa\x96\x04\x15\x52\x6d\x15\x6e\x87\xb9\x3e\x57\xf1\x71\x02\x88" +
	"\x39\xe9\x21\x32\xb3\x3f\x9e\x42\x76\xab\x51\xad\xca\x5d\x0d\x67" +
	"\x87\x07\x9d\xad\xdd\xe8\x00\x59\xe2\x1d\x00\x00\x00\x00\x49\x45" +
	"\x4e\x44\xae\x42\x60\x82"

var errorImage image.Image

func loadError() {
	// Use this image to indicate failed loading of an actual image.
	var err os.Error
	errorImage, err = png.Decode(strings.NewReader(errorImageData))
	if err != nil {
		panic("Unable to load hardcoded error image: " + err.String())
	}
}

var onceError sync.Once

func ErrorImage() image.Image {
	onceError.Do(loadError)
	return errorImage
}

type TranslateGraphics struct {
	Vec   image.Point
	Inner Graphics
}

func (self *TranslateGraphics) Bounds() image.Rectangle { return self.Inner.Bounds() }

func (self *TranslateGraphics) At(x, y int) image.Color {
	return self.Inner.At(x-self.Vec.X, y-self.Vec.Y)
}

func (self *TranslateGraphics) ColorModel() image.ColorModel {
	return self.Inner.ColorModel()
}

func (self *TranslateGraphics) Set(x, y int, c image.Color) {
	self.Inner.Set(x-self.Vec.X, y-self.Vec.Y, c)
}

func (self *TranslateGraphics) Blit(img image.Image, x, y int) {
	self.Inner.Blit(img, x-self.Vec.X, y-self.Vec.Y)
}

func (self *TranslateGraphics) FillRect(rect image.Rectangle, color image.Color) {
	self.Inner.FillRect(rect.Sub(self.Vec), color)
}

func (self *TranslateGraphics) SetClip(clipRect image.Rectangle) {
	self.Inner.SetClip(clipRect)
}

func (self *TranslateGraphics) ClearClip() { self.Inner.ClearClip() }

// Center sets the translation so that the given rectangle will be centered on the given point.
func (self *TranslateGraphics) Center(area image.Rectangle, x, y int) {
	self.Vec = image.Pt(area.Min.X+x-area.Dx()/2, area.Min.Y+y-area.Dy()/2)
}

// Line draws a line on the given draw.Image.
func Line(dst draw.Image, p1, p2 image.Point, col image.Color) {
	for o := range geom.Line(geom.Pt2I{p1.X, p1.Y}, geom.Pt2I{p2.X, p2.Y}).Iter() {
		pt := o.(geom.Pt2I)
		dst.Set(pt.X, pt.Y, col)
	}
}

// Line draws a thick line on the given graphics context. Implementation is
// naive and slow.
func ThickLine(dst Graphics, p1, p2 image.Point, col image.Color, thickness int) {
	offset := thickness / 2
	for o := range geom.Line(geom.Pt2I{p1.X, p1.Y}, geom.Pt2I{p2.X, p2.Y}).Iter() {
		pt := o.(geom.Pt2I)
		rect := image.Rect(pt.X-offset, pt.Y-offset, pt.X+offset+1, pt.Y+offset+1)
		dst.FillRect(rect, col)
	}
}
