package main

import "fmt"
import "rand"
import "time"

import . "teratogen"

var currentLevel int = 1

func main() {
	fmt.Print("Welcome to Teratogen.\n");

	rand.Seed(time.Nanoseconds());

	world := NewWorld();

	world.InitLevel(currentLevel);

	// Game logic
	go func() {
		for {

			// Colemak direction pad.

			// Movement is hjklyubn (Colemak equivalent) move, with bn
			// shifted to nm to keep things on one side on a
			// ergonomic split keyboard.

			GetUISync();
			key := GetKey();
			// When key pressed, clear the message buffer.
			MarkMsgLinesSeen();

			switch key.Printable {
			case 'q':
				Quit();
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
			case 'd':
				Msg("You decide to blow up a bit.\n");
				GameOver("died of exploding head syndrome.");
			}

			RunAI();
			ReleaseUISync();
		}
	}();

	MainUILoop();
}
