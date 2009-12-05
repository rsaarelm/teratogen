package teratogen

import "fmt"

import "libtcod"
import . "gamelib"

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
	result.getch = make(chan KeyEvent);
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

func MainUILoop() {
	con := ui.con;
	world := GetWorld();

	for ui.running {
		con.Clear();

		world.Draw();

		for i := ui.oldestLineSeen; i < GetMsg().NumLines(); i++ {
			con.Print(0, 42 + (i - ui.oldestLineSeen), GetMsg().GetLine(i));
		}

		con.Print(41, 0, fmt.Sprintf("Strength: %v",
			Capitalize(LevelDescription(world.GetPlayer().Strength))));
		con.Print(41, 1, fmt.Sprintf("%v",
			Capitalize(world.GetPlayer().WoundDescription())));

		con.Flush();

		if evt, ok := <-con.Events(); ok {
			switch e := evt.(type) {
			case *KeyEvent:
				ui.getch <- *e;
			}
		}
	}
}