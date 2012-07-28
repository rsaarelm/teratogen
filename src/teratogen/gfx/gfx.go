package gfx

import (
	"image"
)

type Surface32Bit interface {
	Pixels32() []uint32
	Pitch32() int
	Size() image.Point
}
