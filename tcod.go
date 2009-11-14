package tcod

// XXX: C interface has trouble with returned structs. Maybe it's specific to
// the bit-level stuff in TCOD_key_t. Get rid of wrappers when FFI can handle
// TCOD_key_t directly.

// TODO: Return more keypress info than just the char. Keycode and modifier
// key states too.

/*
#include <stdlib.h>
#include "libtcod.h"

static int check_for_keypress(void) {
  TCOD_key_t key;
  key = TCOD_console_check_for_keypress(TCOD_KEY_PRESSED);
  return key.c;
}

void TCOD_console_flush(void);

static TCOD_color_t make_color(uint8 r, uint8 g, uint8 b) {
  TCOD_color_t result;
  result.r = r;
  result.g = g;
  result.b = b;
  return result;
}

static void print_left(int x, int y, TCOD_bkgnd_flag_t flag, const char *txt) {
 TCOD_console_print_left(NULL, x, y, flag, "%s", txt);
}

*/
import "C"
import "unsafe"

type KeyT struct {
	vk C.int;
	c C.char;
	_ uint8;
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

func CheckForKeypress() int {
	return int(C.check_for_keypress());
}

func Flush() {
	C.TCOD_console_flush();
}

func MakeColor(r, g, b uint8) *Color {
	ret := new(Color);
	ret.tcodColor = C.make_color(C.uint8(r), C.uint8(g), C.uint8(b));
	return ret;
}

func PrintLeft(x, y int, bkg BkgndFlag, fmt string) {
	c_fmt := C.CString(fmt);
	C.print_left(C.int(x), C.int(y), C.TCOD_bkgnd_flag_t(bkg), c_fmt);
	C.free(unsafe.Pointer(c_fmt));
}

func SetForeColor(color *Color) {
	C.TCOD_console_set_foreground_color(nil, color.tcodColor);
}
