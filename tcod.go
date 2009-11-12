package tcod

// XXX: C interface has trouble with returned structs. Maybe it's specific to
// the bit-level stuff in TCOD_key_t. Get rid of wrappers when FFI can handle
// TCOD_key_t directly.

// TODO: Return more keypress info than just the char. Keycode and modifier
// key states too.

/*
#include <stdlib.h>
#include "libtcod.h"

int check_for_keypress(void) {
  TCOD_key_t key;
  key = TCOD_console_check_for_keypress(TCOD_KEY_PRESSED);
  return key.c;
}

*/
import "C"
import "unsafe"

type KeyT struct {
	vk C.int;
	c C.char;
	_ uint8;
}

func Init(w int, h int, title string) {
	c_title := C.CString(title);
	C.TCOD_console_init_root(C.int(w), C.int(h), c_title, 0);
	C.free(unsafe.Pointer(c_title));
}

func SetChar(x int, y int, c int) {
	C.TCOD_console_set_char(nil, C.int(x), C.int(y), C.int(c))
}

func CheckForKeypress() int {
	return int(C.check_for_keypress());
}