package gfx

import (
	"image"
	"image/color"
)

type Surface32Bit interface {
	Pixels32() []uint32
	Pitch32() int
	Size() image.Point

	// MapColor converts color data into the internal format of the surface.
	MapColor(c color.Color) uint32

	// GetColor converts an internal color representation into a Color struct.
	GetColor(c32 uint32) color.Color
}
