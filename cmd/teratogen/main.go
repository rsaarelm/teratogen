package main

import (
//	"fmt"
//	"hyades/fs"
	"hyades/num"
)

var currentLevel int = 1

func main() {
//	arch, err := fs.ArchiveFromTarGzFile(fs.SelfExe())
//	if err != nil {
//		fmt.Printf("Self exe archive error: %v\n", err)
//	} else {
//		files, _ := arch.ListFiles()
//		for _, name := range files {
//			fmt.Println(name)
//		}
//	}

	num.RngSeedFromClock()

	InitUI()
	InitMedia()

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
