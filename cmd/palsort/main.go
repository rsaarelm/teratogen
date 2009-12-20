// Construct color slide palettes for hand-painting pixel graphics from
// images.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os"
)

type colorVal uint32

func codeColor(col image.Color) colorVal {
	r, g, b, _ := col.RGBA()
	return colorVal((r >> 24) + ((g >> 24) << 8) + ((b >> 24) << 16))
}

func decodeColor(code colorVal) image.Color {
	return image.RGBAColor{
		byte(code % 0x100),
		byte((code >> 8) % 0x100),
		byte((code >> 16) % 0x100),
		0xff,
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: palsort [SRC_IMAGE] [PAL_IMAGE]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func die(format string, a ...) {
	fmt.Fprintf(os.Stderr, format+"\n", a)
	os.Exit(2)
}

func dieOnErr(err os.Error, format string, a ...) {
	if err != nil {
		die(format, a)
	}
}

func LoadImage(filename string) image.Image {
	data, err := ioutil.ReadFile(filename)
	dieOnErr(err, "Couldn't read file '%s'", filename)
	img, err := png.Decode(bytes.NewBuffer(data))
	dieOnErr(err, "Error loading PNG")
	return img
}

func BuildHistogram(img image.Image) (result map[colorVal]int) {
	result = make(map[colorVal]int)
	for x := 0; x < img.Width(); x++ {
		for y := 0; y < img.Height(); y++ {
			col := codeColor(img.At(x, y))
			prev, _ := result[col]
			result[col] = prev + 1
		}
	}
	return
}

func main() {
	flag.Parse()
	if flag.NArg() < 2 {
		usage()
	}
	img := LoadImage(flag.Arg(0))
	hist := BuildHistogram(img)
	for col, num := range hist {
		fmt.Println(col, num)
	}
}
