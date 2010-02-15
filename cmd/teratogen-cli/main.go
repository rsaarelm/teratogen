package main

import (
	"bufio"
	"fmt"
	"hyades/geom"
	"gostak"
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

	interp := gostak.NewGostakState()
	interp.LoadBuiltins()

	PrintArea()

	Repl(interp)
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

func Repl(interp *gostak.GostakState) {
	for {
		fmt.Print("> ")
		str := ReadLine()
		err := interp.ParseString(str)
		if err != nil {
			fmt.Println("Error:", err)
		}

		for i := 0; i < interp.Len(); i++ {
			fmt.Printf("%d: %v\n", i, interp.At(i))
		}
	}
}
