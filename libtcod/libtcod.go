package libtcod

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

static unpacked_tcod_key_t check_for_keypress(void) {
 return unpack_key_t(TCOD_console_check_for_keypress(TCOD_KEY_PRESSED));
}

static void print_left(int x, int y, TCOD_bkgnd_flag_t flag, const char *txt) {
 TCOD_console_print_left(NULL, x, y, flag, "%s", txt);
}

*/
import "C"
import "unsafe"

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

type KeyT struct {
	Vk Keycode;
	C byte;
	Pressed bool;
	Lalt bool;
	Lctrl bool;
	Ralt bool;
	Rctrl bool;
	Shift bool;
}

type BkgndFlag int
const (
	BkgndNone = iota;
	BkgndSet;
	BkgndMultiply;
	BkgndLighten;
	BkgndDarken;
	BkgndScreen;
	BkgndColorDodge;
	BkgndColorBurn;
	BkgndAdd;
	BkgndAddAlpha;
	BkgndBurn;
	BkgndOverlay;
	BkgndAlpha;
)

type Color struct {
	tcodColor C.TCOD_color_t;
}

func Init(w int, h int, title string) {
	c_title := C.CString(title);
	C.TCOD_console_init_root(C.int(w), C.int(h), c_title, 0);
	C.free(unsafe.Pointer(c_title));
}

func PutChar(x int, y int, c int, bkg BkgndFlag) {
	C.TCOD_console_put_char(
		nil, C.int(x), C.int(y), C.int(c), C.TCOD_bkgnd_flag_t(bkg))
}

// TODO: Return a keypress struct instead of char value only.
func CheckForKeypress() int {
	key := C.check_for_keypress();
	return int(key.c);
}

func Flush() {
	C.TCOD_console_flush();
}

func MakeColor(r, g, b uint8) (ret *Color) {
	ret = new(Color);
	ret.tcodColor = C.make_color(C.uint8(r), C.uint8(g), C.uint8(b));
	return;
}

// TODO: varargs for the string, Sprintf for TCOD.
func PrintLeft(x, y int, bkg BkgndFlag, fmt string) {
	c_fmt := C.CString(fmt);
	C.print_left(C.int(x), C.int(y), C.TCOD_bkgnd_flag_t(bkg), c_fmt);
	C.free(unsafe.Pointer(c_fmt));
}

func SetForeColor(color *Color) {
	C.TCOD_console_set_foreground_color(nil, color.tcodColor);
}

func Clear() {
	C.TCOD_console_clear(nil);
}