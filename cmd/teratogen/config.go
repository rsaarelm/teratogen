package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"hyades/babble"
	"hyades/fs"
	"hyades/keyboard"
	"hyades/num"
	"io/ioutil"
	"json"
	"os"
	"path"
	"teratogen/game"
	"unsafe"
)

type Config struct {
	Sound       bool
	Fullscreen  bool
	KeyLayout   string
	RngSeed     string
	Scale       int
	TileScale   int
	ArchiveFile string
}

func DefaultConfig() *Config { return &Config{false, false, "qwerty", "", 2, 2, fs.SelfExe()} }

var config *Config

func usage() {
	fmt.Fprintf(os.Stderr, "usage: teratogen [OPTION]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func configFileName() string {
	return path.Join(os.Getenv("HOME"), ".teratogenrc")
}

func loadRcFile(config *Config) os.Error {
	if data, err := ioutil.ReadFile(configFileName()); err == nil {
		jsonErr := json.Unmarshal(data, config)
		if jsonErr != nil {
			return jsonErr
		}
	} else {
		// Didn't find config file, fail silently.
	}
	return nil
}

func ParseConfig() {
	config = DefaultConfig()

	versionQuery := false

	if jsonErr := loadRcFile(config); jsonErr != nil {
		fmt.Fprintf(os.Stderr, "Config error in %s: %s\n", configFileName(), jsonErr)
	}

	flag.BoolVar(&config.Fullscreen, "fullscreen", config.Fullscreen, "Run in full screen mode")
	flag.StringVar(&config.KeyLayout, "layout", config.KeyLayout, "Keyboard layout: qwerty|dvorak|colemak")
	flag.StringVar(&config.RngSeed, "logos", config.RngSeed, "Genesis seed")
	flag.IntVar(&config.Scale, "scale", config.Scale, "Window scaling factor, 1|2")
	flag.IntVar(&config.TileScale, "tilescale", config.TileScale, "Tile scaling factor, 1|2")
	flag.StringVar(&config.ArchiveFile, "archive", fs.SelfExe(), "Media archive file")
	flag.BoolVar(&versionQuery, "version", false, "Print version and exit")
	flag.Usage = usage
	flag.Parse()

	if versionQuery {
		fmt.Println(game.Version)
		os.Exit(0)
	}

	// XXX: If config file had bad values, these are presented as command line
	// argument errors, not config file errors.

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

	screenWidth = baseScreenWidth
	screenHeight = baseScreenHeight

	FontW = baseFontW
	FontH = baseFontH

	TileW = baseTileW * config.TileScale
	TileH = baseTileH * config.TileScale
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
