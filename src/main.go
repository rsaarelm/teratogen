package main

import "fmt"
import "rand"
import "time"

import . "gamelib"
import . "teratogen"

var currentLevel int = 1

func main() {
	fmt.Print("Welcome to Teratogen.\n");
	running := true;
	getch := make(chan byte);
	oldestLineSeen := 0;

	rand.Seed(time.Nanoseconds());

	world := NewWorld();

	world.InitLevel(currentLevel);

	// Game logic
	go func() {
		for {
			key := <-getch;
			// When key pressed, clear the message buffer.
			oldestLineSeen = GetMsg().NumLines() - 1;

			// Colemak direction pad.

			// Movement is hjklyubn (Colemak equivalent) move, with bn
			// shifted to nm to keep things on one side on a
			// ergonomic split keyboard.

			switch key {
			case 'q':
				running = false;
			case 'u':
				SmartMovePlayer(0);
			case 'y':
				SmartMovePlayer(1);
			case 'i':
				SmartMovePlayer(2);
			case '.':
				SmartMovePlayer(3);
			case ',':
				SmartMovePlayer(4);
			case 'm':
				SmartMovePlayer(5);
			case 'n':
				SmartMovePlayer(6);
			case 'l':
				SmartMovePlayer(7);
			case 'p':
				Msg("Some text for the buffer...\n");
			}

			RunAI();
		}
	}();

	for running {
		GetConsole().Clear();

		world.Draw();

		for i := oldestLineSeen; i < GetMsg().NumLines(); i++ {
			GetConsole().Print(0, 42 + (i - oldestLineSeen), GetMsg().GetLine(i));
		}

		GetConsole().Print(41, 0, fmt.Sprintf("Strength: %v",
			Capitalize(LevelDescription(world.GetPlayer().Strength))));
		GetConsole().Print(41, 1, fmt.Sprintf("%v",
			Capitalize(world.GetPlayer().WoundDescription())));

		GetConsole().Flush();

		if evt, ok := <-GetConsole().Events(); ok {
			switch e := evt.(type) {
			case *KeyEvent:
				getch <- byte(e.Printable);
			}
		}
	}
}
