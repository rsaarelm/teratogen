package main

import "fmt"
import "rand"
import "time"

import "libtcod"
import . "fomalhaut"
import . "teratogen"

func movePlayerDir(world *World, dir int) {
	world.ClearLosSight();
	world.MoveCreature(world.GetPlayer(), Dir8ToVec(dir));
	world.DoLos(world.GetPlayer().GetPos());
}

func smartMove(world *World, dir int) {
	player := world.GetPlayer();
	vec := Dir8ToVec(dir);
	target := player.GetPos().Plus(vec);

	for ent := range world.EntitiesAt(target) {
		if world.IsEnemyOf(player, ent) {
			world.Attack(player, ent);
			return;
		}
	}
	// No attack, move normally.
	movePlayerDir(world, dir);
}

func main() {
	fmt.Print("Welcome to Teratogen.\n");
	running := true;
	getch := make(chan byte);

	rand.Seed(time.Nanoseconds());

	//libtcod.Init(80, 50, "Teratogen");
	Con = NewConsole(libtcod.NewLibtcodConsole(80, 50, "Teratogen"));

	Msg = NewMsgOut();

	world := NewWorld();

	world.InitLevel(1);

	world.DoLos(world.GetPlayer().GetPos());

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
			case 'u':
				smartMove(world, 0);
			case 'y':
				smartMove(world, 1);
			case 'i':
				smartMove(world, 2);
			case '.':
				smartMove(world, 3);
			case ',':
				smartMove(world, 4);
			case 'm':
				smartMove(world, 5);
			case 'n':
				smartMove(world, 6);
			case 'l':
				smartMove(world, 7);
			case 'p':
				fmt.Fprint(Msg, "Some text for the buffer... ");
			}
		}
	}();

	for running {
		Con.Clear();

		world.Draw();

		Con.Print(0, 0, Msg.GetLine());
		Con.Print(0, 42, fmt.Sprintf("Strength: %v",
			Capitalize(LevelDescription(world.GetPlayer().Strength))));
		Con.Print(24, 42, fmt.Sprintf("%v",
			Capitalize(WoundDescription(world.GetPlayer().Wounds))));

		Con.Flush();

		if evt, ok := <-Con.Events(); ok {
			switch e := evt.(type) {
			case *KeyEvent:
				getch <- byte(e.Printable);
			}
		}
	}
}
