package main

import (
	"fmt"
	"hyades/num"
//	"os"
	"teratogen/teratogen"
)

func main() {
	teratogen.ParseConfig()

	seed := num.RandStateFromClock()
/*
	var err os.Error
	if config.RngSeed != "" {
		seed, err = teratogen.BabbleToRandState(config.RngSeed)
		if err != nil {
			fmt.Printf("Invalid genesis seed: %s.\n", config.RngSeed)
			seed = num.RandStateFromClock()
		}
	}
*/

	num.RestoreRngState(seed)
	fmt.Println("Logos:", teratogen.RandStateToBabble(seed))

//	teratogen.InitUI()
//	teratogen.InitMedia()

	teratogen.InitWorld()

	teratogen.GetWorld().InitLevel(1)

	go teratogen.LogicLoop()
//	teratogen.MainUILoop()
}
