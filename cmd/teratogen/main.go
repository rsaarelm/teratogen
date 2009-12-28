package main

import (
	"hyades/dbg"
	"hyades/num"
	"hyades/txt"
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

	keymap := txt.KeyMap(txt.ColemakMap)

	// Game logic
	go func() {
		for {
			// Colemak direction pad.

			// Movement is hjklyubn (Colemak equivalent) move, with bn
			// shifted to nm to keep things on one side on a
			// ergonomic split keyboard.

			GetUISync()
			key := keymap.Map(GetKey())
			// When key pressed, clear the message buffer.
			MarkMsgLinesSeen()

			switch key {
			case 'q':
				Quit()
			case 'a':
				AnimTest()
			case 'k':
				SmartMovePlayer(0)
			case 'u':
				SmartMovePlayer(1)
			case 'l':
				SmartMovePlayer(2)
			case 'n':
				SmartMovePlayer(3)
			case 'j':
				SmartMovePlayer(4)
			case 'b':
				SmartMovePlayer(5)
			case 'h':
				SmartMovePlayer(6)
			case 'y':
				SmartMovePlayer(7)
			case ',':
				SmartPlayerPickup()
			case 'i':
				// Show inventory.
				Msg("Carried:")
				first := true
				item := world.GetPlayer().GetChild()
				for item != nil {
					if first {
						first = false
						Msg(" %v", item.Name)
					} else {
						Msg(", %v", item.Name)
					}
					item = item.GetSibling()
				}
				if first {
					Msg(" nothing.\n")
				} else {
					Msg(".\n")
				}
			case 'd':
				// Drop item.
				// XXX: No selection UI yet, just drop the first one in inventory.
				item := world.GetPlayer().GetChild()
				if item != nil {
					item.RemoveSelf()
					Msg("Dropped %v.\n", item.GetName())
				} else {
					Msg("Nothing to drop.\n")
				}
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
