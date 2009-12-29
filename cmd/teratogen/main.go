package main

import (
	"exp/iterable"
	"hyades/dbg"
	"hyades/keyboard"
	"hyades/num"
	"os"
	"strings"
)

var currentLevel int = 1

func PlayerInput() {
loop: for {
		key := keymap.Map(GetKey())
		MarkMsgLinesSeen()
		// When key pressed, clear the message buffer.

		switch key {
		case '.':
			// Idle.
			break loop
		case 'q':
			Quit()
		case 'a':
			AnimTest()
		case 'k', keyboard.K_UP, keyboard.K_KP8:
			SmartMovePlayer(0)
			break loop
		case 'u', keyboard.K_PAGEUP, keyboard.K_KP9:
			SmartMovePlayer(1)
			break loop
		case 'l', keyboard.K_RIGHT, keyboard.K_KP6:
			SmartMovePlayer(2)
			break loop
		case 'n', keyboard.K_PAGEDOWN, keyboard.K_KP3:
			SmartMovePlayer(3)
			break loop
		case 'j', keyboard.K_DOWN, keyboard.K_KP2:
			SmartMovePlayer(4)
			break loop
		case 'b', keyboard.K_END, keyboard.K_KP1:
			SmartMovePlayer(5)
			break loop
		case 'h', keyboard.K_LEFT, keyboard.K_KP4:
			SmartMovePlayer(6)
			break loop
		case 'y', keyboard.K_HOME, keyboard.K_KP7:
			SmartMovePlayer(7)
			break loop
		case 'z':
			words := strings.Split("If the egg's shell does not break, the chick will die without being born. We are the chick; the egg is the world. If the world's shell does not break, we will die without being born. Break the world's shell! For the sake of revolutionizing the world!", " ", 0)
			cvt := make([]interface{}, len(words))
			for i, a := range words {
				cvt[i] = a
			}
			x, ok := MultiChoiceDialogA("Select word.", cvt)
			if ok {
				Msg("%d: %v\n", x, cvt[x])
			} else {
				Msg("Okay then.\n")
			}
		case ',':
			if SmartPlayerPickup() != nil {
				break loop
			}
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
		case 'e':
			EquipMenu()
		case 'd':
			// Drop item.
			player := world.GetPlayer()
			if player.HasContents() {
				item, ok := ObjectChoiceDialog(
					"Drop which item?", iterable.Data(player.Contents()))
				if ok {
					item := item.(*Entity)
					item.RemoveSelf()
					Msg("Dropped %v.\n", item.GetName())
					break loop
				} else {
					Msg("Okay, then.\n")
				}
			} else {
				Msg("Nothing to drop.\n")
			}
		case '>':
			PlayerEnterStairs()
			break loop
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
	}
}

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
			PlayerInput()

			RunAI()
			ReleaseUISync()
		}
	}()

	MainUILoop()
}
