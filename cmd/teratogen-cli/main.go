package main

import (
	"bufio"
	"fmt"
	"hyades/geom"
	"hyades/num"
	"os"
	game "teratogen"
)

func main() {
	seed := num.RandStateFromClock()
	fmt.Println("Logos:", game.RandStateToBabble(seed))

	context := game.NewContext()
	context.InitGame()
	fmt.Println("Command-line teratogen.")

	PrintArea()

	fmt.Print("> ")
	fmt.Println(ReadLine())
}

func PrintArea() {
	los := game.GetLos()
	area := game.GetArea()

	for y := 0; y < area.Height(); y++ {
	Cell: for x := 0; x < area.Width(); x++ {
			pt := geom.Pt2I{x, y}
			if los.Get(pt) == game.LosSeen {
				for i := range game.EntitiesAt(pt).Iter() {
					// XXX: Prints the non-@ character if an item is on the same
					// cell as the player and shows up before the player in the
					// iteration.
					if i.(*game.Blob) == game.GetPlayer() {
						fmt.Print("@")
					} else {
						fmt.Print("x")
					}
					continue Cell
				}
				terrainType := area.GetTerrain(pt)
				if game.IsObstacleTerrain(terrainType) {
					fmt.Print("#")
				} else {
					fmt.Print(".")
				}
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
}

func ReadLine() (result string) {
	in := bufio.NewReader(os.Stdin)
	result, err := in.ReadString('\n')
	if err != nil {
		panic("ReadLine error.")
	}
	return
}
