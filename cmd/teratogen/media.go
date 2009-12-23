package main

import (
	"bytes"
	"exp/draw"
	"fmt"
	"hyades/dbg"
	"hyades/fs"
	"hyades/gfx"
	"image"
	"image/png"
	"once"
	"os"
)

// Media server package

var archive *fs.Archive

var cache map[string]interface{}

var transparentColor = image.RGBAColor{0xff, 0x00, 0xff, 0xff}

func Load(filename string) (data []byte, err os.Error) {
	once.Do(initArchive)
	return archive.ReadFile(filename)
}

func initArchive() {
	cache = make(map[string]interface{})
	arch, err := fs.ArchiveFromTarGzFile(fs.SelfExe())
	dbg.AssertNil(err, "%v", err)
	archive = arch
}

func makeTiles(basename string, filename string, width, height int) (result []image.Image) {
	data, err := Load(filename)
	dbg.AssertNil(err, "%v", err)
	png, err := png.Decode(bytes.NewBuffer(data))
	dbg.AssertNil(err, "%v", err)
	sheet := gfx.DoubleScaleImage(png)
	tiles := gfx.MakeTiles(sheet, gfx.DefaultConstructor, width, height)
	result = make([]image.Image, len(tiles))
	for i, tile := range tiles {
		result[i] = ui.context.Convert(tile.(image.Image))
		gfx.FilterTransparent(result[i].(draw.Image), transparentColor)
		id := fmt.Sprintf("%v:%v", basename, i)
		cache[id] = result[i]
	}
	return
}

func InitMedia() {
	once.Do(initArchive)
	makeTiles("font", "media/font.png", TileW, TileH)
	makeTiles("guys", "media/chars.png", TileW, TileH)
	makeTiles("tiles", "media/tiles.png", TileW, TileH)
	makeTiles("items", "media/items_1.png", TileW, TileH)
}

func Media(name string) interface{} { return cache[name] }
