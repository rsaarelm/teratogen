package main

import (
	"bytes"
	"fmt"
	. "hyades/common"
	"hyades/fs"
	"hyades/sdl"
	"once"
	"os"
)

// Media server package

var archive *fs.Archive

var cache map[string] interface{}

func Load(filename string) (data []byte, err os.Error) {
	once.Do(initArchive)
	return archive.ReadFile(filename)
}

func initArchive() {
	cache = make(map[string] interface{})
	arch, err := fs.ArchiveFromTarGzFile(fs.SelfExe())
	AssertNil(err, "%v", err)
	archive = arch
}

func makeTiles(basename string,
	filename string,
	width, height, xoff, yoff, xgap, ygap int) (result []*sdl.Surface) {
	data, err := Load(filename)
	AssertNil(err, "%v", err)
	sheet, err := sdl.MakePngSurface(bytes.NewBuffer(data))
	AssertNil(err, "%v", err)
	result = sheet.MakeTiles(width, height, xoff, yoff, xgap, ygap)
	sheet.FreeSurface()
	for i, x := range result {
		cache[fmt.Sprintf("%v:%v", basename, i)] = x
	}
	return
}

func InitMedia() {
	once.Do(initArchive)
	makeTiles("font", "media/font.png", 8, 8, 0, 0, 0, 0)
	makeTiles("guys", "media/chars_1.png", 8, 8, 0, 0, 0, 0)
	makeTiles("tiles", "media/tiles_2.png", 8, 8, 0, 0, 0, 0)
}

func Media(name string) interface{} {
	return cache[name]
}
