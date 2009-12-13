package main

import (
	"hyades/console"
	"hyades/geom"
)

func ConsoleClear(con console.ConsoleBase) {
	w, h := con.GetDim()
	for pt := range geom.PtIter(0, 0, w, h) {
		con.Set(pt.X, pt.Y, ' ', console.RGB{0, 0, 0}, console.RGB{0, 0, 0})
	}
}

func ConsolePrint(con console.ConsoleBase, x, y int, txt string, fore, back console.RGB) {
	for i := 0; i < len(txt); i++ {
		con.Set(x+i, y, int(txt[i]), fore, back)
	}
}
