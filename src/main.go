package main

import "fmt"
import "math"
import "rand"
import "time"

import "libtcod"
import . "fomalhaut"
import . "teratogen"

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

func smartMove(world *World, dir int) {
	player := world.GetPlayer();
	vec := dir8ToVec(dir);
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

type MsgOut struct {
	tickerLine string;
	input chan string;
}

func NewMsgOut() (result *MsgOut) {
	result = new(MsgOut);
	result.input = make(chan string);
	go result.runTicker();
	return;
}

func (self *MsgOut) runTicker() {
	for {
		const tickerWidth = 80;
		const lettersAtTime = 1;
		const letterDelayNs = 1e9 * 0.20;

		if append, ok := <-self.input; ok {
			self.tickerLine = self.tickerLine + append;
		}

		// XXX: lettesDelayNs doesn't evaluate to an exact integer due
		// to rounding errors, and casting inexact floats to integers
		// is a compile-time error, so we need an extra Floor
		// operation here.
		time.Sleep(int64(math.Floor(letterDelayNs) * lettersAtTime));
		for x := 0; x <= lettersAtTime; x++ {
			self.tickerLine = updateTicker(self.tickerLine, tickerWidth);
		}
	}
}

func (self *MsgOut) GetLine() string { return self.tickerLine; }

func (self *MsgOut) WriteString(str string) {
	self.input <- str;
}

// TODO: MsgOut io.Writer implemetation.

func main() {
	fmt.Print("Welcome to Teratogen.\n");
	running := true;
	getch := make(chan byte);

	rand.Seed(time.Nanoseconds());

	libtcod.Init(80, 50, "Teratogen");

	world := NewWorld();

	world.InitLevel(1);

	world.DoLos(world.GetPlayer().GetPos());

	msg := NewMsgOut();

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
				smartMove(world, 0);
			case 'l':
				smartMove(world, 1);
			case 'i':
				smartMove(world, 2);
			case 'k':
				smartMove(world, 3);
			case 'n':
				smartMove(world, 4);
			case 'b':
				smartMove(world, 5);
			case 'h':
				smartMove(world, 6);
			case 'j':
				smartMove(world, 7);
			case 'p':
				msg.WriteString("Some text for the buffer... ");
			}
		}
	}();

	for running {
		libtcod.Clear();

		world.Draw();
		libtcod.SetForeColor(libtcod.MakeColor(192, 192, 192));
		libtcod.PrintLeft(0, 0, libtcod.BkgndNone, msg.GetLine());

		libtcod.Flush();

		key := libtcod.CheckForKeypress();
		if key != 0 {
			getch <- byte(key);
		}
	}
}
