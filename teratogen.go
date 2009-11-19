package main

import "fmt"
import "math"
import "time"

import "tcod"
import "fomalhaut"
import "sync"

const tickerWidth = 80;

func updateTicker(str string, lineLength int) string {
	return fomalhaut.PadString(fomalhaut.EatPrefix(str, 1), lineLength);
}

type World struct {
	PlayerX, PlayerY int;
	Lock *sync.RWMutex;
}

func MakeWorld() (result *World) {
	result = new(World);
	result.PlayerX = 40;
	result.PlayerY = 20;
	result.Lock = new(sync.RWMutex);
	return;
}

// TODO: Event system for changing world, event handler does lock/unlock, all
// changes in events. "Transactional database".

func (self *World) MovePlayer(dx, dy int) {
	self.Lock.Lock();
	defer self.Lock.Unlock();

	self.PlayerX += dx;
	self.PlayerY += dy;
}

func main() {
	fmt.Print("Welcome to Teratogen.\n");
	running := true;
	getch := make(chan byte);

	tcod.Init(80, 50, "Teratogen");
	tcod.SetForeColor(tcod.MakeColor(255, 255, 0));
	tcod.PutChar(0, 0, 64, tcod.BkgndNone);
	tcod.PrintLeft(0, 2, tcod.BkgndNone, "Hello, world!");
	tcod.SetForeColor(tcod.MakeColor(255, 0, 0));
	tcod.PutChar(0, 0, 65, tcod.BkgndNone);
	tcod.Flush();
	world := MakeWorld();

	tickerLine := "                                                                                Teratogen online. ";

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

	tcod.SetForeColor(tcod.MakeColor(0, 255, 0));
	for running {
		tcod.Clear();
		tcod.SetForeColor(tcod.MakeColor(192, 192, 192));
		tcod.PrintLeft(0, 0, tcod.BkgndNone, tickerLine);
		tcod.SetForeColor(tcod.MakeColor(0, 255, 0));

		tcod.PutChar(world.PlayerX, world.PlayerY, '@', tcod.BkgndNone);
		tcod.Flush();

		key := tcod.CheckForKeypress();
		if key != 0 {
			getch <- byte(key);
		}
	}
}
