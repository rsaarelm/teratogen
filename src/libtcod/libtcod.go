package libtcod

import . "fomalhaut"

// XXX: C interface has trouble with returned structs. Maybe it's specific to
// the bit-level stuff in TCOD_key_t. Get rid of wrappers when FFI can handle
// TCOD_key_t directly.

// TODO: Return more keypress info than just the char. Keycode and modifier
// key states too.

/*
#include <stdlib.h>
#include "libtcod.h"

void TCOD_console_flush(void);

static TCOD_color_t make_color(uint8 r, uint8 g, uint8 b) {
  TCOD_color_t result;
  result.r = r;
  result.g = g;
  result.b = b;
  return result;
}

// Golang can't handle structs with 1-bit fields, so we make a normalized
// version of the TCOD_key_t struct.
typedef struct {
  TCOD_keycode_t vk;
  char c;
  char pressed;
  char lalt;
  char lctrl;
  char ralt;
  char rctrl;
  char shift;
} unpacked_tcod_key_t;

static unpacked_tcod_key_t unpack_key_t(TCOD_key_t key) {
  unpacked_tcod_key_t result;
  result.vk = key.vk;
  result.c = key.c;
  result.pressed = key.pressed;
  result.lalt = key.lalt;
  result.lctrl = key.lctrl;
  result.ralt = key.ralt;
  result.rctrl = key.rctrl;
  result.shift = key.shift;
  return result;
}

static unpacked_tcod_key_t wait_for_keypress(bool flush) {
 return unpack_key_t(TCOD_console_wait_for_keypress(flush));
}

*/
import "C"
import "unsafe"

func NewLibtcodConsole(w, h int, title string) (ConsoleBase) {
	result := new (libtcodConsole);

	// Init libtcod.
	c_title := C.CString(title);
	C.TCOD_console_init_root(C.int(w), C.int(h), c_title, 0);
	C.free(unsafe.Pointer(c_title));

	result.eventChannel = make(chan ConsoleEvent);

	spawnEventListeners(result.eventChannel);

	return result;
}

type libtcodConsole struct {
	eventChannel chan ConsoleEvent;
}

func (self *libtcodConsole) Set(
	x, y int, symbol int, foreColor, backColor RGB) {
	C.TCOD_console_set_foreground_color(nil, C.make_color(
		C.uint8(foreColor[0]), C.uint8(foreColor[1]), C.uint8(foreColor[2])));
	C.TCOD_console_set_background_color(nil, C.make_color(
		C.uint8(backColor[0]), C.uint8(backColor[1]), C.uint8(backColor[2])));
	C.TCOD_console_put_char(
		nil, C.int(x), C.int(y), C.int(symbol), C.TCOD_bkgnd_flag_t(0));
}

func (self *libtcodConsole) Get(
	x, y int) (symbol int, foreColor, backColor RGB) {
	symbol = int(C.TCOD_console_get_char(nil, C.int(x), C.int(y)));
	foreColor = unpackColor(
		C.TCOD_console_get_fore(nil, C.int(x), C.int(y)));
	backColor = unpackColor(
		C.TCOD_console_get_back(nil, C.int(x), C.int(y)));
	return;
}

func (self *libtcodConsole) Events() <-chan ConsoleEvent {
	return self.eventChannel;
}

func (self *libtcodConsole) GetDim() (width, height int) {
	width = int(C.TCOD_console_get_width(nil));
	height = int(C.TCOD_console_get_height(nil));
	return;
}

func (self *libtcodConsole) ColorsDiffer(col1, col2 RGB) bool {
	// Full-color console.
	return true;
}

func (self *libtcodConsole) ShowCursorAt(x, y int) {
	// TODO
}

func (self *libtcodConsole) HideCursor() {
	// TODO
}

func (self *libtcodConsole) Flush() {
	C.TCOD_console_flush();
}

func unpackColor(tcodColor C.TCOD_color_t) RGB {
	return RGB{byte(tcodColor.r), byte(tcodColor.g), byte(tcodColor.b)};
}

func spawnEventListeners(events chan<- ConsoleEvent) {
	keyListener := func() {
		for {
			key := C.wait_for_keypress(C.bool(0));
			event := &KeyEvent{int(key.vk), int(key.c), key.pressed != 0};
			events <- event;
		}
	};

	go keyListener();

	// TODO mouse, resize, quit listeners.
}

// Obsolete stuff below.
type Keycode byte
const (
	TCODK_NONE = iota;
	TCODK_ESCAPE;
	TCODK_BACKSPACE;
	TCODK_TAB;
	TCODK_ENTER;
	TCODK_SHIFT;
	TCODK_CONTROL;
	TCODK_ALT;
	TCODK_PAUSE;
	TCODK_CAPSLOCK;
	TCODK_PAGEUP;
	TCODK_PAGEDOWN;
	TCODK_END;
	TCODK_HOME;
	TCODK_UP;
	TCODK_LEFT;
	TCODK_RIGHT;
	TCODK_DOWN;
	TCODK_PRINTSCREEN;
	TCODK_INSERT;
	TCODK_DELETE;
	TCODK_LWIN;
	TCODK_RWIN;
	TCODK_APPS;
	TCODK_0;
	TCODK_1;
	TCODK_2;
	TCODK_3;
	TCODK_4;
	TCODK_5;
	TCODK_6;
	TCODK_7;
	TCODK_8;
	TCODK_9;
	TCODK_KP0;
	TCODK_KP1;
	TCODK_KP2;
	TCODK_KP3;
	TCODK_KP4;
	TCODK_KP5;
	TCODK_KP6;
	TCODK_KP7;
	TCODK_KP8;
	TCODK_KP9;
	TCODK_KPADD;
	TCODK_KPSUB;
	TCODK_KPDIV;
	TCODK_KPMUL;
	TCODK_KPDEC;
	TCODK_KPENTER;
	TCODK_F1;
	TCODK_F2;
	TCODK_F3;
	TCODK_F4;
	TCODK_F5;
	TCODK_F6;
	TCODK_F7;
	TCODK_F8;
	TCODK_F9;
	TCODK_F10;
	TCODK_F11;
	TCODK_F12;
	TCODK_NUMLOCK;
	TCODK_SCROLLLOCK;
	TCODK_SPACE;
	TCODK_CHAR;
)
