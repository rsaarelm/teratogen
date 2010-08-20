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
	"os"
	"sync"
	"time"
)

// Media server package

var archive *fs.Archive

var cache map[string]interface{}

var onceMedia sync.Once

func Load(filename string) (data []byte, err os.Error) {
	onceMedia.Do(initArchive)
	return archive.ReadFile(filename)
}

func initArchive() {
	cache = make(map[string]interface{})
	arch, err := fs.ArchiveFromTarGzFile(config.ArchiveFile)
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

func PlaySound(name string) {
	if config.Sound {
		cache[name].(sfx.Sound).Play()
	}
}

func loadFonts() {
	fontfile, err := Load("media/04round_bold.ttf")
	dbg.AssertNoError(err)

	ui.font, err = ui.context.LoadFont(fontfile, 8)
	dbg.AssertNoError(err)
}

func InitMedia() {
	fmt.Printf("Initializing media... ")
	defer fmt.Printf("Done.\n")
	onceMedia.Do(initArchive)
	makeTiles("font", "media/font.png", FontW, FontH, 1)
	makeTiles("chars", "media/chars.png", TileW, TileH, config.TileScale)
	makeTiles("tiles", "media/tiles.png", TileW, TileH, config.TileScale)
	makeTiles("items", "media/items.png", TileW, TileH, config.TileScale)
	if config.Sound {
		makeSounds()
	}

	loadFonts()
}

func Media(name string) interface{} { return cache[name] }

func SaveScreenshot() {
	GetUISync()
	screen := ui.context.SdlScreen()

	shot := image.NewRGBA(screen.Bounds().Dx(), screen.Bounds().Dy())
	draw.Draw(shot, screen.Bounds(), screen, image.ZP)

	// XXX: Alpha must be removed from the image or the shot bitmap goes funny.
	gfx.FilterImage(shot, gfx.OpaqueAlphaFn)

	filename := fmt.Sprintf("/tmp/sshot-%d.png", time.UTC().Seconds())
	file, err := os.Open(filename, os.O_WRONLY|os.O_CREAT, 0666)

	if err == nil {
		png.Encode(file, shot)
	}

	file.Close()
	ReleaseUISync()
}
