package main

import "fmt"
import "math"
import "rand"
import "time"

import "libtcod"
import . "fomalhaut"
import . "teratogen"

const tickerWidth = 80;

func updateTicker(str string, lineLength int) string {
	return PadString(EatPrefix(str, 1), lineLength);
}

func main() {
	fmt.Print("Welcome to Teratogen.\n");
	running := true;
	getch := make(chan byte);

	rand.Seed(time.Nanoseconds());

	libtcod.Init(80, 50, "Teratogen");

	world := NewWorld();

	// TODO: Player entity struct
	var player Entity;
	player = nil;

	world.InitLevel(1, player);

	tickerLine := "";

	go func() {
		for {
			const lettersAtTime = 1;
			const letterDelayNs = 1e9 * 0.20;
			// XXX: lettesDelayNs doesn't evaluate to an exact
			// integer due to rounding errors, and casting inexact
			// floats to integers is a compile-time error, so we
			// need an extra Floor operation here.
			time.Sleep(int64(math.Floor(letterDelayNs) * lettersAtTime));
			for x := 0; x <= lettersAtTime; x++ {
				tickerLine = updateTicker(tickerLine, tickerWidth);
			}
		}
	}();

	// Game logic
	go func() {
		for {
			key := <-getch;
			switch key {
			case 'q':
				running = false;
				// Colemak direction pad.
			case 'n':
				world.MovePlayer(-1, 0);
			case ',':
				world.MovePlayer(0, 1);
			case 'i':
				world.MovePlayer(1, 0);
			case 'u':
				world.MovePlayer(0, -1);
			case 'p':
				tickerLine += "Some text for the buffer... ";
			}
		}
	}();

	for running {
		libtcod.Clear();

		world.Draw();
		libtcod.SetForeColor(libtcod.MakeColor(192, 192, 192));
		libtcod.PrintLeft(0, 0, libtcod.BkgndNone, tickerLine);

		libtcod.Flush();

		key := libtcod.CheckForKeypress();
		if key != 0 {
			getch <- byte(key);
		}
	}
}
