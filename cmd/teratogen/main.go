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

func LogicLoop() {
	for {
		GetUISync()
		game.DoTurn()
		ReleaseUISync()
	}
}
