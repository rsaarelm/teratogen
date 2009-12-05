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

	sync := make(chan int, 1);

	// Game logic
	go func() {
		for {
			key := <-Getch();

			// When key pressed, clear the message buffer.
			MarkMsgLinesSeen();

			// Colemak direction pad.

			// Movement is hjklyubn (Colemak equivalent) move, with bn
			// shifted to nm to keep things on one side on a
			// ergonomic split keyboard.

			<-sync;

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
			}

			RunAI();
			sync <- 1;
		}
	}();

	sync <- 1;
	MainUILoop(sync);
}
