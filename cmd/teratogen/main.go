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
	return "/tmp/saved.gam"
}

// Try to load a save if there is one. Delete the successfully loaded save.
func JourneyOnward() bool {
	GetUISync()
	defer ReleaseUISync()
	fileName := SaveFileName()
	loadFile, err := os.Open(fileName, os.O_RDONLY, 0666)
	if err != nil {
		return false
	}
	game.GetContext().Deserialize(loadFile)
	loadFile.Close()
	os.Remove(fileName)
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
