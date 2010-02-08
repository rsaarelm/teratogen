package teratogen

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"hyades/babble"
	"hyades/keyboard"
	"hyades/num"
	"os"
	"unsafe"
)

type Config struct {
	Sound      bool
	Fullscreen bool
	KeyLayout  string
	RngSeed    string
	Scale      int
	TileScale  int
}

func DefaultConfig() *Config { return &Config{false, false, "qwerty", "", 2, 1} }

var config *Config

func usage() {
	fmt.Fprintf(os.Stderr, "usage: teratogen [OPTION]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func ParseConfig() {
	config = DefaultConfig()

	flag.BoolVar(&config.Sound, "sound", config.Sound, "Play sounds.")
	flag.BoolVar(&config.Fullscreen, "fullscreen", config.Fullscreen, "Run in full screen mode.")
	flag.StringVar(&config.KeyLayout, "layout", config.KeyLayout, "Keyboard layout: qwerty|dvorak|colemak.")
	flag.StringVar(&config.RngSeed, "logos", config.RngSeed, "Genesis seed.")
	flag.IntVar(&config.Scale, "scale", config.Scale, "Window scaling factor, 1|2")
	flag.IntVar(&config.TileScale, "tilescale", config.TileScale, "Tile scaling factor, 1|2")
	flag.Usage = usage
	flag.Parse()

	switch config.KeyLayout {
	case "qwerty":
		keymap = keyboard.KeyMap(keyboard.QwertyMap)
	case "colemak":
		keymap = keyboard.KeyMap(keyboard.ColemakMap)
	case "dvorak":
		keymap = keyboard.KeyMap(keyboard.DvorakMap)
	default:
		usage()
	}

	if config.Scale < 1 || config.Scale > 2 {
		usage()
	}

	if config.TileScale < 1 || config.TileScale > 2 {
		usage()
	}

	screenWidth = config.Scale * baseScreenWidth
	screenHeight = config.Scale * baseScreenHeight

	FontW = config.Scale * baseFontW
	FontH = config.Scale * baseFontH

	TileW = baseTileW * config.Scale * config.TileScale
	TileH = baseTileH * config.Scale * config.TileScale
}

func RandStateToBabble(state num.RandState) string {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, state)
	return babble.EncodeToString(buf.Bytes())
}

func BabbleToRandState(bab string) (result num.RandState, err os.Error) {
	data, err := babble.DecodeString(bab)
	if err != nil {
		return
	}
	if len(data) != unsafe.Sizeof(num.RandState(0)) {
		err = os.NewError("Bad babble data length.")
		return
	}
	var state num.RandState
	err = binary.Read(bytes.NewBuffer(data), binary.BigEndian, &state)
	result = state
	return
}
