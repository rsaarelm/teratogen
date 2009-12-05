package main

import "fmt"
import "rand"
import "time"

import . "gamelib"
import . "teratogen"

var currentLevel int = 1

func movePlayerDir(dir int) {
	world := GetWorld();
	world.ClearLosSight();
	world.GetPlayer().TryMove(Dir8ToVec(dir));
	world.DoLos(world.GetPlayer().GetPos());
}

func smartMove(dir int) {
	world := GetWorld();
	player := world.GetPlayer();
	vec := Dir8ToVec(dir);
	target := player.GetPos().Plus(vec);

	for ent := range world.EntitiesAt(target) {
		if IsEnemyOf(player, ent) {
			Attack(player, ent);
			return;
		}
	}
	// No attack, move normally.
	movePlayerDir(dir);
}

func RunAI() {
	world := GetWorld();
	enemyCount := 0;
	for crit := range world.IterCreatures() {
		if crit != world.GetPlayer() { enemyCount++; }
		DoAI(crit);
	}

	// Go to next level when all creatures are killed.
	// TODO: Show message, get keypress, before flipping to the next level.
	if enemyCount == 0 {
		currentLevel++;
		world.InitLevel(currentLevel);
	}
}

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
				smartMove(0);
			case 'y':
				smartMove(1);
			case 'i':
				smartMove(2);
			case '.':
				smartMove(3);
			case ',':
				smartMove(4);
			case 'm':
				smartMove(5);
			case 'n':
				smartMove(6);
			case 'l':
				smartMove(7);
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
