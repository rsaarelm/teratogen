package fomalhaut

// Minimal features for implementing a console.
type ConsoleBase interface {
	Set(x, y int, symbol int, foreColor, backColor ConsoleColor);
        Get(x, y int) (symbol int, foreColor, backColor ConsoleColor);
        Events() <-chan ConsoleEvent;
	GetDim() (width, height int);
        EncodeColor(r, g, b byte) ConsoleColor;
        DecodeColor(col ConsoleColor) (r, g, b byte);
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

