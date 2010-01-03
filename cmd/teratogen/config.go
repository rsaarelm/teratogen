package main

import (
	"flag"
	"fmt"
	"hyades/keyboard"
	"os"
)

type Config struct {
	Sound      bool
	Fullscreen bool
	KeyLayout  string
	RngSeed    int64
	TileScale  int
}

func DefaultConfig() *Config { return &Config{false, false, "qwerty", -1, 2} }

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
	flag.Int64Var(&config.RngSeed, "seed", config.RngSeed, "Random number generator seed.")
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
