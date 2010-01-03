// Keyboard input and key symbol utilities.

package keyboard

type KeyMap string

// XXX: DvorakMap not tested for typos.
const (
	ColemakMap KeyMap = " !\"#$%&'()*+,-./0123456789Pp<=>?@ABCGKETHLYNUMJ:RQSDFIVWXOZ[\\]^_`abcgkethlynumj;rqsdfivwxoz{|}~"
	DvorakMap  KeyMap = " !Q#$%&q()*}w'e[0123456789ZzW]E{@ANIHDYUJGCVPMLSRXO:KF><BT?/\\=^\"`anihdyujgcvpmlsrxo;kf.,bt/_|+~"
	QwertyMap  KeyMap = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
)

func (self KeyMap) Map(key int) int {
	if key-32 < len(self) {
		return int(self[key-32])
	}
	return key
}

// Special key flags.
const (
	Nonprintable = 1 << (16 + iota)
	LShift
	RShift
	LCtrl
	RCtrl
	LAlt
	RAlt

	Shift = LShift | RShift
	Ctrl  = LCtrl | RCtrl
	Alt   = LAlt | RAlt

	// Use this bit mask to strip modifier information from a keypress value.
	StripModifiers = 0x1ffff
)

// Key symbols, using libSDL keysym convention for names and values.
const (
	K_BACKSPACE   = Nonprintable | 8
	K_TAB         = Nonprintable | 9
	K_CLEAR       = Nonprintable | 12
	K_RETURN      = Nonprintable | 13
	K_PAUSE       = Nonprintable | 19
	K_ESCAPE      = Nonprintable | 27
	K_KP0         = Nonprintable | 256
	K_KP1         = Nonprintable | 257
	K_KP2         = Nonprintable | 258
	K_KP3         = Nonprintable | 259
	K_KP4         = Nonprintable | 260
	K_KP5         = Nonprintable | 261
	K_KP6         = Nonprintable | 262
	K_KP7         = Nonprintable | 263
	K_KP8         = Nonprintable | 264
	K_KP9         = Nonprintable | 265
	K_KP_PERIOD   = Nonprintable | 266
	K_KP_DIVIDE   = Nonprintable | 267
	K_KP_MULTIPLY = Nonprintable | 268
	K_KP_MINUS    = Nonprintable | 269
	K_KP_PLUS     = Nonprintable | 270
	K_KP_ENTER    = Nonprintable | 271
	K_KP_EQUALS   = Nonprintable | 272
	K_UP          = Nonprintable | 273
	K_DOWN        = Nonprintable | 274
	K_RIGHT       = Nonprintable | 275
	K_LEFT        = Nonprintable | 276
	K_INSERT      = Nonprintable | 277
	K_HOME        = Nonprintable | 278
	K_END         = Nonprintable | 279
	K_PAGEUP      = Nonprintable | 280
	K_PAGEDOWN    = Nonprintable | 281
	K_F1          = Nonprintable | 282
	K_F2          = Nonprintable | 283
	K_F3          = Nonprintable | 284
	K_F4          = Nonprintable | 285
	K_F5          = Nonprintable | 286
	K_F6          = Nonprintable | 287
	K_F7          = Nonprintable | 288
	K_F8          = Nonprintable | 289
	K_F9          = Nonprintable | 290
	K_F10         = Nonprintable | 291
	K_F11         = Nonprintable | 292
	K_F12         = Nonprintable | 293
	K_F13         = Nonprintable | 294
	K_F14         = Nonprintable | 295
	K_F15         = Nonprintable | 296
	K_NUMLOCK     = Nonprintable | 300
	K_CAPSLOCK    = Nonprintable | 301
	K_SCROLLOCK   = Nonprintable | 302
	K_RSHIFT      = Nonprintable | 303
	K_LSHIFT      = Nonprintable | 304
	K_RCTRL       = Nonprintable | 305
	K_LCTRL       = Nonprintable | 306
	K_RALT        = Nonprintable | 307
	K_LALT        = Nonprintable | 308
	K_RMETA       = Nonprintable | 309
	K_LMETA       = Nonprintable | 310
	K_LSUPER      = Nonprintable | 311
	K_RSUPER      = Nonprintable | 312
	K_MODE        = Nonprintable | 313
	K_COMPOSE     = Nonprintable | 314
	K_HELP        = Nonprintable | 315
	K_PRINT       = Nonprintable | 316
	K_SYSREQ      = Nonprintable | 317
	K_BREAK       = Nonprintable | 318
	K_MENU        = Nonprintable | 319
	K_POWER       = Nonprintable | 320
	K_EURO        = Nonprintable | 321
	K_UNDO        = Nonprintable | 322
)
