package main

import (
	"fmt"
	"hyades/num"
	"os"
	game "teratogen"
)

func main() {
	InitialConfig()
	InitUI()
	InitMedia()

	fx = new(SdlEffects)
	game.InitEffects(fx)
	game.NewContext().InitGame()

	JourneyOnward()
	go LogicLoop()
	MainUILoop()
}

var fx *SdlEffects

func InitialConfig() {
	ParseConfig()

	seed := num.RandStateFromClock()
	var err os.Error
	if config.RngSeed != "" {
		seed, err = BabbleToRandState(config.RngSeed)
		if err != nil {
			fmt.Printf("Invalid genesis seed: %s.\n", config.RngSeed)
			seed = num.RandStateFromClock()
		}
	}

	num.RestoreRngState(seed)
	fmt.Println("Logos:", RandStateToBabble(seed))
}

func SaveFileName() string {
	return "teratogen.sav"
}

// Try to load a save if there is one. Delete the successfully loaded save.
func JourneyOnward() bool {
	GetUISync()
	defer ReleaseUISync()
	err := game.LoadGame(SaveFileName())
	if err != nil {
		// XXX: Since we got a global var for game state, undo the damage from a
		// botched load by reiniting the game state.

		// XXX: The reinited game state will have a different rng seed than the
		// one reported at the command line.
		game.NewContext().InitGame()

		return false
	}
	os.Remove(SaveFileName())
	game.Msg("Game loaded.\n")
	return true
}

func LogicLoop() {
	for {
		GetUISync()
		game.DoTurn()
		ReleaseUISync()
	}
}
