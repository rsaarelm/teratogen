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

func dir8ToVec(dir int) Vec2I {
	switch dir {
	case 0: return Vec2I{0, -1};
	case 1: return Vec2I{1, -1};
	case 2: return Vec2I{1, 0};
	case 3: return Vec2I{1, 1};
	case 4: return Vec2I{0, 1};
	case 5: return Vec2I{-1, 1};
	case 6: return Vec2I{-1, 0};
	case 7: return Vec2I{-1, -1};
	}
	panic("Invalid dir");
}

func movePlayerDir(world *World, dir int) {
	world.ClearLosSight();
	world.MovePlayer(dir8ToVec(dir));
	world.DoLos(world.GetPlayer().GetPos());
}

func main() {
	fmt.Print("Welcome to Teratogen.\n");
	running := true;
	getch := make(chan byte);

	rand.Seed(time.Nanoseconds());

	libtcod.Init(80, 50, "Teratogen");

	world := NewWorld();

	world.InitLevel(1);

	world.DoLos(world.GetPlayer().GetPos());

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
			// Colemak direction pad.

			// Movement is hjklyubn (Colemak equivalent) move, with bn
			// shifted to nm to keep things on one side on a
			// ergonomic split keyboard.

			switch key {
			case 'q':
				running = false;
			case 'e':
				movePlayerDir(world, 0);
			case 'l':
				movePlayerDir(world, 1);
			case 'i':
				movePlayerDir(world, 2);
			case 'm':
				movePlayerDir(world, 3);
			case 'n':
				movePlayerDir(world, 4);
			case 'k':
				movePlayerDir(world, 5);
			case 'h':
				movePlayerDir(world, 6);
			case 'j':
				movePlayerDir(world, 7);
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
