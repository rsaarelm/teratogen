package gfx

import (
	"fmt"
	"testing"
	"image"
)

func match(t *testing.T, testName string, expt interface{}, got interface{}) {
	exptStr, gotStr := fmt.Sprint(expt), fmt.Sprint(got)
	if exptStr != gotStr {
		t.Errorf("%s: Expected %s, got %s", testName, exptStr, gotStr)
	}
}

func TestStringRGB(t *testing.T) {
	match(t, "StringRGB", "#ffeedd", StringRGB(image.RGBAColor{0xff, 0xee, 0xdd, 0xff}))
	match(t, "StringRGB", "#000000", StringRGB(image.RGBAColor{0x00, 0x00, 0x00, 0xff}))
	match(t, "StringRGB", "#030201", StringRGB(image.RGBAColor{0x03, 0x02, 0x01, 0xff}))
	match(t, "StringRGB", "#030200", StringRGB(image.RGBAColor{0x03, 0x02, 0x00, 0xff}))

	match(t, "StringRGBA", "#ffeeddcc", StringRGBA(image.RGBAColor{0xff, 0xee, 0xdd, 0xcc}))
	match(t, "StringRGBA", "#00000000", StringRGBA(image.RGBAColor{0x00, 0x00, 0x00, 0x00}))
	match(t, "StringRGBA", "#03020105", StringRGBA(image.RGBAColor{0x03, 0x02, 0x01, 0x05}))
	match(t, "StringRGBA", "#03020000", StringRGBA(image.RGBAColor{0x03, 0x02, 0x00, 0x00}))
}

func makeCol(t *testing.T, desc string, expt [4]byte) {
	col, err := MakeColor(desc)
	if err != nil {
		t.Errorf("Failed to make color '%s': %s", desc, err.String())
		return
	}
	gotStr := StringRGB(col)
	exptStr := StringRGB(image.RGBAColor{expt[0], expt[1], expt[2], expt[3]})
	if gotStr != exptStr {
		match(t, "MakeColor", exptStr, gotStr)
	}
}

func TestMakeColor(t *testing.T) {
	makeCol(t, "#fff", [4]byte{0xff, 0xff, 0xff, 0xff})
	makeCol(t, "#ffff", [4]byte{0xff, 0xff, 0xff, 0xff})
	makeCol(t, "#ffffff", [4]byte{0xff, 0xff, 0xff, 0xff})
	makeCol(t, "#ffffffff", [4]byte{0xff, 0xff, 0xff, 0xff})

	makeCol(t, "#123", [4]byte{0x11, 0x22, 0x33, 0xff})
	makeCol(t, "#1234", [4]byte{0x11, 0x22, 0x33, 0x44})
	makeCol(t, "#010203", [4]byte{0x01, 0x02, 0x03, 0xff})
	makeCol(t, "#01020304", [4]byte{0x01, 0x02, 0x03, 0x04})
}
