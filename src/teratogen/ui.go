package teratogen

import "fmt"

import "libtcod"
import . "gamelib"

type UI struct {
	getch chan KeyEvent;
	msg *MsgOut;

	con *Console
}

var ui = newUI();

func newUI() (result *UI) {
	result = new(UI);
	result.getch = make(chan KeyEvent);
	result.msg = NewMsgOut();
	result.con = NewConsole(libtcod.NewLibtcodConsole(80, 50, "Teratogen"));

	return;
}

func GetConsole() *Console { return ui.con; }

func GetMsg() *MsgOut { return ui.msg; }

func Msg(format string, a ...) {
	fmt.Fprintf(ui.msg, format, a);
}