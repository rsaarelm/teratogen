package gfx

import (
	"fmt"
	"image"
	"os"
)

func StringRGB(col image.Color) string {
	if col == nil { return "<nil>" }
	r, g, b, _ := col.RGBA()
	return fmt.Sprintf("#%02x%02x%02x",
		byte(r>>24),
		byte(g>>24),
		byte(b>>24))
}

func StringRGBA(col image.Color) string {
	if col == nil { return "<nil>" }
	r, g, b, a := col.RGBA()
	return fmt.Sprintf("#%02x%02x%02x%02x",
		byte(r>>24),
		byte(g>>24),
		byte(b>>24),
		byte(a>>24))
}

func MakeColor(desc string) (col image.Color, err os.Error) {
	err = os.NewError("TBD")
	return
}
