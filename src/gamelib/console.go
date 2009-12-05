package gamelib

// This is a wrapper class for consoles which implements complex display
// logic. It holds a reference to a minimal implementation object which does
// the implementation-specific things.
type Console struct {
	impl ConsoleBase;
	fore, back RGB;
}

func NewConsole(impl ConsoleBase) (result *Console) {
	DieIfNil(impl, "console implementation");
	result = new(Console);
	result.impl = impl;
	result.fore = RGB{255, 255, 255};
	result.back = RGB{0, 0, 0};
	return;
}

func (self *Console) Clear() {
	w, h := self.impl.GetDim();
	for pt := range PtIter(0, 0, w, h) {
		self.impl.Set(pt.X, pt.Y, ' ', self.fore, self.back);
	}
}

func (self *Console) SetCFB(x, y int, symbol int, foreColor, backColor RGB) {
	self.impl.Set(x, y, symbol, foreColor, backColor);
}

func (self *Console) SetC(x, y int, symbol int) {
	_, foreColor, backColor := self.Get(x, y);
	self.SetCFB(x, y, symbol, foreColor, backColor);
}

func (self *Console) SetF(x, y int, foreColor RGB) {
	symbol, _, backColor := self.Get(x, y);
	self.SetCFB(x, y, symbol, foreColor, backColor);
}

func (self *Console) SetB(x, y int, backColor RGB) {
	symbol, foreColor, _ := self.Get(x, y);
	self.SetCFB(x, y, symbol, foreColor, backColor);
}

func (self *Console) SetCF(x, y int, symbol int, foreColor RGB) {
	_, _, backColor := self.Get(x, y);
	self.SetCFB(x, y, symbol, foreColor, backColor);
}

func (self *Console) SetFB(x, y int, foreColor, backColor RGB) {
	symbol, _, _ := self.Get(x, y);
	self.impl.Set(x, y, symbol, foreColor, backColor);
}

func (self *Console) Get(x, y int) (symbol int, foreColor, backColor RGB) {
	return self.impl.Get(x, y);
}

func (self *Console) SetForeCol(col RGB) { self.fore = col; }

func (self *Console) SetBackCol(col RGB) { self.back = col; }

func (self *Console) Flush() { self.impl.Flush(); }

func (self *Console) Events() <-chan ConsoleEvent { return self.impl.Events(); }

func (self *Console) Print(x, y int, txt string) {
	for i := 0; i < len(txt); i++ {
		self.impl.Set(x + i, y, int(txt[i]), self.fore, self.back);
	}
}

func (self *Console) ForeColorsDiffer(col1, col2 RGB) bool {
	return self.impl.ForeColorsDiffer(col1, col2);
}

func (self *Console) BackColorsDiffer(col1, col2 RGB) bool {
	return self.impl.BackColorsDiffer(col1, col2);
}

// TODO: Canonical keycode enumeration. Use the ones from SDL.

// Minimal features for implementing a console.
type ConsoleBase interface {
	Set(x, y int, symbol int, foreColor, backColor RGB);
        Get(x, y int) (symbol int, foreColor, backColor RGB);
        Events() <-chan ConsoleEvent;
	GetDim() (width, height int);
	// Return whether the console is able to differentiate between the two
	// colors. Different for foreground and background, because the
	// standard curses console has less background color resolution.
	ForeColorsDiffer(col1, col2 RGB) bool;
	BackColorsDiffer(col1, col2 RGB) bool;
        ShowCursorAt(x, y int);
        HideCursor();
        Flush();
}

type ConsoleEvent interface {}

type KeyEvent struct {
	Code int;
        Printable int;
	// TODO: KeyDown / KeyUp instead of just keypress?
	// TODO: Modifier buttons
}

type MouseEvent struct {
	Action MouseAction;
        X, Y int;
        Buttons [8]bool;
}

type ResizeEvent struct {
	Width, Height int;
}

type QuitEvent struct {}

type MouseAction byte const (
	MouseDown = iota;
        MouseUp;
        MouseMove;
)
