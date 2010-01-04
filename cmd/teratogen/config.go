package main

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
	TileScale  int
}

func DefaultConfig() *Config { return &Config{false, false, "qwerty", "", 2} }

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
	flag.StringVar(&config.RngSeed, "seed", config.RngSeed, "Random number generator seed.")
	flag.IntVar(&config.TileScale, "scale", config.TileScale, "Tile scaling factor, 1|2|3|4")
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

	if config.TileScale < 1 || config.TileScale > 4 {
		usage()
	}
	TileW = 8 * config.TileScale
	TileH = 8 * config.TileScale
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
