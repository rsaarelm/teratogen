package main

import . "hyades/gamelib"

func ConsoleClear(console ConsoleBase) {
	w, h := console.GetDim();
	for pt := range PtIter(0, 0, w, h) {
		console.Set(pt.X, pt.Y, ' ', RGB{0, 0, 0}, RGB{0, 0, 0});
	}
}

func ConsolePrint(console ConsoleBase, x, y int, txt string, fore, back RGB) {
	for i := 0; i < len(txt); i++ {
		console.Set(x + i, y, int(txt[i]), fore, back);
	}
}
