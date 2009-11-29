package teratogen

import "math"
import "os"
import "time"

import . "fomalhaut"

var Msg *MsgOut;

func updateTicker(str string, lineLength int) string {
	return PadString(EatPrefix(str, 1), lineLength);
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

func (self *MsgOut) Write(p []byte) (n int, err os.Error) {
	self.input <- string(p);
	n = len(p);
	return;
}