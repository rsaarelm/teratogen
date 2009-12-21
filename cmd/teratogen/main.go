package main

import (
	"hyades/dbg"
	"hyades/num"
	"os"
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

			switch key {
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
				// Experimental save/load
			case 'S':
				saveFile, err := os.Open("/tmp/saved.gam", os.O_WRONLY|os.O_CREAT, 0666)
				dbg.AssertNoError(err)
				world.Serialize(saveFile)
				saveFile.Close()
				Msg("Game saved.\n")
			case 'L':
				loadFile, err := os.Open("/tmp/saved.gam", os.O_RDONLY, 0666)
				if err != nil {
					Msg("Error loading game: " + err.String())
					break
				}
				world = new(World)
				SetWorld(world)
				world.Deserialize(loadFile)
				Msg("Game loaded.\n")
			}

			RunAI()
			ReleaseUISync()
		}
	}()

	MainUILoop()
}
