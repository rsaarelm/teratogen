package main

import (
	"bytes"
	"exp/draw"
	"fmt"
	"hyades/dbg"
	"hyades/fs"
	"hyades/gfx"
	"hyades/sfx"
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

func makeTiles(basename string, filename string, width, height int, scale int) (result []image.Image) {
	data, err := Load(filename)
	dbg.AssertNil(err, "%v", err)
	png, err := png.Decode(bytes.NewBuffer(data))
	dbg.AssertNil(err, "%v", err)
	sheet := gfx.IntScaleImage(png, scale, scale)
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

func makeSounds() {
	healWave := sfx.AmpFilter(0.3, sfx.ADSRFilter(0.0, 0.0, 0.9, 0.2, 0.1,
		sfx.MakeWave(1500.0, sfx.Jump(0.1, 50, sfx.Sawtooth))))

	snd, err := sfx.MonoWaveToSound(ui.context, healWave, 1.0)
	dbg.AssertNoError(err)
	cache["heal"] = snd

	hitWave := sfx.AmpFilter(0.3, sfx.ADSRFilter(0.0, 0.0, 0.9, 0.0, 0.2,
		sfx.MakeWave(300.0, sfx.Noise)))
	snd, err = sfx.MonoWaveToSound(ui.context, hitWave, 1.0)
	dbg.AssertNoError(err)
	cache["hit"] = snd

	deathWave := sfx.AmpFilter(0.4, sfx.ADSRFilter(0.0, 0.0, 0.9, 0.0, 0.8,
		sfx.MakeWave(250.0, sfx.Slide(-200.0, 0.0, 50.0, sfx.Noise))))
	snd, err = sfx.MonoWaveToSound(ui.context, deathWave, 1.0)
	dbg.AssertNoError(err)
	cache["death"] = snd

}

func PlaySound(name string) { cache[name].(sfx.Sound).Play() }

func InitMedia() {
	once.Do(initArchive)
	makeTiles("font", "media/font.png", TileW, TileH, 2)
	makeTiles("chars", "media/chars.png", TileW, TileH, 2)
	makeTiles("tiles", "media/tiles.png", TileW, TileH, 2)
	makeTiles("items", "media/items.png", TileW, TileH, 2)
	makeSounds()
}

func Media(name string) interface{} { return cache[name] }
