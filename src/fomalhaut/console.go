package fomalhaut

// TODO: Canonical keycode enumeration. Use the ones from SDL.

// Minimal features for implementing a console.
type ConsoleBase interface {
	Set(x, y int, symbol int, foreColor, backColor RGB);
        Get(x, y int) (symbol int, foreColor, backColor RGB);
        Events() <-chan ConsoleEvent;
	GetDim() (width, height int);
	// Return whether the console is able to differentiate between the two
	// colors.
	ColorsDiffer(col1, col2 RGB) bool;
        ShowCursorAt(x, y int);
        HideCursor();
        Flush();
}

type ConsoleColor uint32

type ConsoleEvent interface {}

type KeyEvent struct {
	Code int;
        Printable int;
	// True if key pressed, false if key released.
        Pressed bool;
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

// Global console handle

var Console ConsoleBase