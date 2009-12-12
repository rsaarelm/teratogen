package main

import (
	"bytes"
	"compress/gzip"
	"container/vector"
	"fmt"
	. "hyades/gamelib"
	"io/ioutil"
)

var currentLevel int = 1

func main() {
	findArchive()
	RngSeedFromClock()

	InitUI()

	world := NewWorld()

	world.InitLevel(currentLevel)

	// Game logic
	go func() {
		for {

			// Colemak direction pad.

			// Movement is hjklyubn (Colemak equivalent) move, with bn
			// shifted to nm to keep things on one side on a
			// ergonomic split keyboard.

			GetUISync()
			key := GetKey()
			// When key pressed, clear the message buffer.
			MarkMsgLinesSeen()

			switch key.Printable {
			case 'q':
				Quit()
			case 'u':
				SmartMovePlayer(0)
			case 'y':
				SmartMovePlayer(1)
			case 'i':
				SmartMovePlayer(2)
			case '.':
				SmartMovePlayer(3)
			case ',':
				SmartMovePlayer(4)
			case 'm':
				SmartMovePlayer(5)
			case 'n':
				SmartMovePlayer(6)
			case 'l':
				SmartMovePlayer(7)
			case 'p':
				Msg("Some text for the buffer...\n")
			case 'd':
				Msg("You decide to blow up a bit.\n")
				GameOver("died of exploding head syndrome.")
			case '>':
				PlayerEnterStairs()
			case 'c':
				world.ClearLosMapped()
				world.DoLos(world.GetPlayer().GetPos())
				Msg("You feel like you've forgotten something.\n")
			}

			RunAI()
			ReleaseUISync()
		}
	}()

	MainUILoop()
}

func findArchive() {
	// XXX: Way to get self exe that only works on unix-like platforms.
	data, err := ioutil.ReadFile("/proc/self/exe")
	if err != nil {
		Die("Couldn't read self exe.")
	}
	points := new(vector.Vector)

	// Don't write the magic byte as a straight literal, as that'll then
	// show up as an extra magic string in the binary code.
	magic := make([]byte, 3)
	magic[0] = 0x1f
	magic[1] = 0x8b
	magic[2] = 0x08

	// Scan the self exe backwards for the gz identifying magic sequence.
        for i := len(data) - len(magic) - 1; i > 0; i-- {
		found := true
		for j := 0; j < len(magic); j++ {
			if data[i + j] != magic[j] { found = false }
		}
		if found { points.Push(i) }
	}
	fmt.Println("Possible tar archive offsets:")
	for pt := range points.Iter() {
		fmt.Printf("%v: ", pt)
		pos := pt.(int)
		inf, err := gzip.NewInflater(bytes.NewBuffer(data[pos:]))
		if err != nil {
			fmt.Println("Invalid pos.")
		} else {
			_, err2 := inf.Read(make([]byte, 1))
			if err2 != nil {
				fmt.Println("Invalid pos.")
			} else {
				fmt.Println("Looks like a gzip.")
			}
		}
	}
	fmt.Println("End")

}
