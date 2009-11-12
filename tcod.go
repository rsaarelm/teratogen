package tcod

/*
#include <stdlib.h>
#include "libtcod.h"
*/
import "C"
import "unsafe"

func Init(w int, h int, title string) {
	c_title := C.CString(title);
	C.TCOD_console_init_root(C.int(w), C.int(h), c_title, 0);
	C.free(unsafe.Pointer(c_title));
}