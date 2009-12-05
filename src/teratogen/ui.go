package teratogen

import "fmt"
import "time"

import "libtcod"
import . "gamelib"

const redrawIntervalNs = 30e6
const capFps = true

type UI struct {
	getch chan KeyEvent;
	msg *MsgOut;
	con *Console;
	running bool;

	// Show message lines beyond this to player.
	oldestLineSeen int;
}

var ui = newUI();

func newUI() (result *UI) {
	result = new(UI);
	result.getch = make(chan KeyEvent, 16);
	result.msg = NewMsgOut();
	result.con = NewConsole(libtcod.NewLibtcodConsole(80, 50, "Teratogen"));
	result.running = true;

	return;
}

func GetConsole() *Console { return ui.con; }

func GetMsg() *MsgOut { return ui.msg; }

func Msg(format string, a ...) {
	fmt.Fprintf(ui.msg, format, a);
}

func Quit() {
	ui.running = false;
}

func MarkMsgLinesSeen() {
	ui.oldestLineSeen = ui.msg.NumLines() - 1;
}

func Getch() <-chan KeyEvent {
	return ui.getch;
}

func MainUILoop(sync chan int) {
	con := ui.con;

	updater := time.Tick(redrawIntervalNs);

	for ui.running {
		// XXX: Can't put grabbing and releasing sync next to each
		// other, to the very beginning and end of the loop, or the
		// script side will never get sync.

		if capFps {
			// Wait for the next tick before repainting.
			<-updater;
		}

		con.Clear();

		// Synched block which accesses the game world. Don't run
		// scripts during this.
		<-sync;
		world := GetWorld();

		world.Draw();

		for i := ui.oldestLineSeen; i < GetMsg().NumLines(); i++ {
			con.Print(0, 42 + (i - ui.oldestLineSeen), GetMsg().GetLine(i));
		}

		con.Print(41, 0, fmt.Sprintf("Strength: %v",
			Capitalize(LevelDescription(world.GetPlayer().Strength))));
		con.Print(41, 1, fmt.Sprintf("%v",
			Capitalize(world.GetPlayer().WoundDescription())));
		sync <- 1;

		con.Flush();

		handleInput();
	}
}

func handleInput() {
	if evt, ok := <-ui.con.Events(); ok {
		switch e := evt.(type) {
		case *KeyEvent:
			bufferKeypress(e);
		}
	}
}

func bufferKeypress(e *KeyEvent) {
	// Non-blocking send.
	ok := ui.getch <- *e;
	if !ok {
		// If the key buffer is full, drop the
		// oldest input and push the new one
		// in.

		// XXX: Possible to lose input here,
		// if another goroutine grabbed the
		// head input between the line above
		// and this.
		<-ui.getch;
		ok2 := ui.getch <- *e;

		Assert(ok2, "Couldn't write to key buffer after dropping a value.");
	}
}