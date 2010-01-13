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
	keyCh := key & StripModifiers
	mods := key &^ keyCh
	if keyCh >= 32 && keyCh-32 < len(self) {
		keyCh = int(self[keyCh-32])
	}
	return keyCh | mods
}

func KeyCode(keyName string) (code int, ok bool) {
	code, ok = keyCodes[keyName]
	return
}

func KeyName(keyCode int) (name string, ok bool) {
	name, ok = keyNames[keyCode]
	return
}

// ExtendedKeyName returns a key name with modifiers
func ExtendedKeyName(keyCode int) (name string, ok bool) {
	if keyCode&Ctrl != 0 {
		name += "C-"
		keyCode &^= Ctrl
	}
	if keyCode&Alt != 0 {
		name += "M-"
		keyCode &^= Alt
	}
	if keyCode&Shift != 0 {
		name += "S-"
		keyCode &^= Shift
	}
	if innerName, ok := KeyName(keyCode); ok {
		return name + innerName, true
	}
	return "", false
}

func KeyCode(keyName string) (code int, ok bool) {
	code, ok = keyCodes[keyName]
	return
}

func KeyName(keyCode int) (name string, ok bool) {
	name, ok = keyNames[keyCode]
	return
}

// ExtendedKeyName returns a key name with modifiers
func ExtendedKeyName(keyCode int) (name string, ok bool) {
	if keyCode&Ctrl != 0 {
		name += "C-"
		keyCode &^= Ctrl
	}
	if keyCode&Alt != 0 {
		name += "M-"
		keyCode &^= Alt
	}
	if keyCode&Shift != 0 {
		name += "S-"
		keyCode &^= Shift
	}
	if innerName, ok := KeyName(keyCode); ok {
		return name + innerName, true
	}
	return "", false
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
	K_BACKSPACE    = Nonprintable | 8
	K_TAB          = Nonprintable | 9
	K_CLEAR        = Nonprintable | 12
	K_RETURN       = Nonprintable | 13
	K_PAUSE        = Nonprintable | 19
	K_ESCAPE       = Nonprintable | 27
	K_SPACE        = 32
	K_EXCLAIM      = 33
	K_QUOTEDBL     = 34
	K_HASH         = 35
	K_DOLLAR       = 36
	K_AMPERSAND    = 38
	K_QUOTE        = 39
	K_LEFTPAREN    = 40
	K_RIGHTPAREN   = 41
	K_ASTERISK     = 42
	K_PLUS         = 43
	K_COMMA        = 44
	K_MINUS        = 45
	K_PERIOD       = 46
	K_SLASH        = 47
	K_0            = 48
	K_1            = 49
	K_2            = 50
	K_3            = 51
	K_4            = 52
	K_5            = 53
	K_6            = 54
	K_7            = 55
	K_8            = 56
	K_9            = 57
	K_COLON        = 58
	K_SEMICOLON    = 59
	K_LESS         = 60
	K_EQUALS       = 61
	K_GREATER      = 62
	K_QUESTION     = 63
	K_AT           = 64
	K_A            = 65
	K_B            = 66
	K_C            = 67
	K_D            = 68
	K_E            = 69
	K_F            = 70
	K_G            = 71
	K_H            = 72
	K_I            = 73
	K_J            = 74
	K_K            = 75
	K_L            = 76
	K_M            = 77
	K_N            = 78
	K_O            = 79
	K_P            = 80
	K_Q            = 81
	K_R            = 82
	K_S            = 83
	K_T            = 84
	K_U            = 85
	K_V            = 86
	K_W            = 87
	K_X            = 88
	K_Y            = 89
	K_Z            = 90
	K_LEFTBRACKET  = 91
	K_BACKSLASH    = 92
	K_RIGHTBRACKET = 93
	K_CARET        = 94
	K_UNDERSCORE   = 95
	K_BACKQUOTE    = 96
	K_a            = 97
	K_b            = 98
	K_c            = 99
	K_d            = 100
	K_e            = 101
	K_f            = 102
	K_g            = 103
	K_h            = 104
	K_i            = 105
	K_j            = 106
	K_k            = 107
	K_l            = 108
	K_m            = 109
	K_n            = 110
	K_o            = 111
	K_p            = 112
	K_q            = 113
	K_r            = 114
	K_s            = 115
	K_t            = 116
	K_u            = 117
	K_v            = 118
	K_w            = 119
	K_x            = 120
	K_y            = 121
	K_z            = 122
	K_DELETE       = Nonprintable | 127
	K_KP0          = Nonprintable | 256
	K_KP1          = Nonprintable | 257
	K_KP2          = Nonprintable | 258
	K_KP3          = Nonprintable | 259
	K_KP4          = Nonprintable | 260
	K_KP5          = Nonprintable | 261
	K_KP6          = Nonprintable | 262
	K_KP7          = Nonprintable | 263
	K_KP8          = Nonprintable | 264
	K_KP9          = Nonprintable | 265
	K_KP_PERIOD    = Nonprintable | 266
	K_KP_DIVIDE    = Nonprintable | 267
	K_KP_MULTIPLY  = Nonprintable | 268
	K_KP_MINUS     = Nonprintable | 269
	K_KP_PLUS      = Nonprintable | 270
	K_KP_ENTER     = Nonprintable | 271
	K_KP_EQUALS    = Nonprintable | 272
	K_UP           = Nonprintable | 273
	K_DOWN         = Nonprintable | 274
	K_RIGHT        = Nonprintable | 275
	K_LEFT         = Nonprintable | 276
	K_INSERT       = Nonprintable | 277
	K_HOME         = Nonprintable | 278
	K_END          = Nonprintable | 279
	K_PAGEUP       = Nonprintable | 280
	K_PAGEDOWN     = Nonprintable | 281
	K_F1           = Nonprintable | 282
	K_F2           = Nonprintable | 283
	K_F3           = Nonprintable | 284
	K_F4           = Nonprintable | 285
	K_F5           = Nonprintable | 286
	K_F6           = Nonprintable | 287
	K_F7           = Nonprintable | 288
	K_F8           = Nonprintable | 289
	K_F9           = Nonprintable | 290
	K_F10          = Nonprintable | 291
	K_F11          = Nonprintable | 292
	K_F12          = Nonprintable | 293
	K_F13          = Nonprintable | 294
	K_F14          = Nonprintable | 295
	K_F15          = Nonprintable | 296
	K_NUMLOCK      = Nonprintable | 300
	K_CAPSLOCK     = Nonprintable | 301
	K_SCROLLOCK    = Nonprintable | 302
	K_RSHIFT       = Nonprintable | 303
	K_LSHIFT       = Nonprintable | 304
	K_RCTRL        = Nonprintable | 305
	K_LCTRL        = Nonprintable | 306
	K_RALT         = Nonprintable | 307
	K_LALT         = Nonprintable | 308
	K_RMETA        = Nonprintable | 309
	K_LMETA        = Nonprintable | 310
	K_LSUPER       = Nonprintable | 311
	K_RSUPER       = Nonprintable | 312
	K_MODE         = Nonprintable | 313
	K_COMPOSE      = Nonprintable | 314
	K_HELP         = Nonprintable | 315
	K_PRINT        = Nonprintable | 316
	K_SYSREQ       = Nonprintable | 317
	K_BREAK        = Nonprintable | 318
	K_MENU         = Nonprintable | 319
	K_POWER        = Nonprintable | 320
	K_EURO         = Nonprintable | 321
	K_UNDO         = Nonprintable | 322
)

var keyCodes = map[string]int{
	"K_BACKSPACE": K_BACKSPACE,
	"K_TAB": K_TAB,
	"K_CLEAR": K_CLEAR,
	"K_RETURN": K_RETURN,
	"K_PAUSE": K_PAUSE,
	"K_ESCAPE": K_ESCAPE,
	"K_SPACE": K_SPACE,
	"K_EXCLAIM": K_EXCLAIM,
	"K_QUOTEDBL": K_QUOTEDBL,
	"K_HASH": K_HASH,
	"K_DOLLAR": K_DOLLAR,
	"K_AMPERSAND": K_AMPERSAND,
	"K_QUOTE": K_QUOTE,
	"K_LEFTPAREN": K_LEFTPAREN,
	"K_RIGHTPAREN": K_RIGHTPAREN,
	"K_ASTERISK": K_ASTERISK,
	"K_PLUS": K_PLUS,
	"K_COMMA": K_COMMA,
	"K_MINUS": K_MINUS,
	"K_PERIOD": K_PERIOD,
	"K_SLASH": K_SLASH,
	"K_0": K_0,
	"K_1": K_1,
	"K_2": K_2,
	"K_3": K_3,
	"K_4": K_4,
	"K_5": K_5,
	"K_6": K_6,
	"K_7": K_7,
	"K_8": K_8,
	"K_9": K_9,
	"K_COLON": K_COLON,
	"K_SEMICOLON": K_SEMICOLON,
	"K_LESS": K_LESS,
	"K_EQUALS": K_EQUALS,
	"K_GREATER": K_GREATER,
	"K_QUESTION": K_QUESTION,
	"K_AT": K_AT,
	"K_A": K_A,
	"K_B": K_B,
	"K_C": K_C,
	"K_D": K_D,
	"K_E": K_E,
	"K_F": K_F,
	"K_G": K_G,
	"K_H": K_H,
	"K_I": K_I,
	"K_J": K_J,
	"K_K": K_K,
	"K_L": K_L,
	"K_M": K_M,
	"K_N": K_N,
	"K_O": K_O,
	"K_P": K_P,
	"K_Q": K_Q,
	"K_R": K_R,
	"K_S": K_S,
	"K_T": K_T,
	"K_U": K_U,
	"K_V": K_V,
	"K_W": K_W,
	"K_X": K_X,
	"K_Y": K_Y,
	"K_Z": K_Z,
	"K_LEFTBRACKET": K_LEFTBRACKET,
	"K_BACKSLASH": K_BACKSLASH,
	"K_RIGHTBRACKET": K_RIGHTBRACKET,
	"K_CARET": K_CARET,
	"K_UNDERSCORE": K_UNDERSCORE,
	"K_BACKQUOTE": K_BACKQUOTE,
	"K_a": K_a,
	"K_b": K_b,
	"K_c": K_c,
	"K_d": K_d,
	"K_e": K_e,
	"K_f": K_f,
	"K_g": K_g,
	"K_h": K_h,
	"K_i": K_i,
	"K_j": K_j,
	"K_k": K_k,
	"K_l": K_l,
	"K_m": K_m,
	"K_n": K_n,
	"K_o": K_o,
	"K_p": K_p,
	"K_q": K_q,
	"K_r": K_r,
	"K_s": K_s,
	"K_t": K_t,
	"K_u": K_u,
	"K_v": K_v,
	"K_w": K_w,
	"K_x": K_x,
	"K_y": K_y,
	"K_z": K_z,
	"K_DELETE": K_DELETE,
	"K_KP0": K_KP0,
	"K_KP1": K_KP1,
	"K_KP2": K_KP2,
	"K_KP3": K_KP3,
	"K_KP4": K_KP4,
	"K_KP5": K_KP5,
	"K_KP6": K_KP6,
	"K_KP7": K_KP7,
	"K_KP8": K_KP8,
	"K_KP9": K_KP9,
	"K_KP_PERIOD": K_KP_PERIOD,
	"K_KP_DIVIDE": K_KP_DIVIDE,
	"K_KP_MULTIPLY": K_KP_MULTIPLY,
	"K_KP_MINUS": K_KP_MINUS,
	"K_KP_PLUS": K_KP_PLUS,
	"K_KP_ENTER": K_KP_ENTER,
	"K_KP_EQUALS": K_KP_EQUALS,
	"K_UP": K_UP,
	"K_DOWN": K_DOWN,
	"K_RIGHT": K_RIGHT,
	"K_LEFT": K_LEFT,
	"K_INSERT": K_INSERT,
	"K_HOME": K_HOME,
	"K_END": K_END,
	"K_PAGEUP": K_PAGEUP,
	"K_PAGEDOWN": K_PAGEDOWN,
	"K_F1": K_F1,
	"K_F2": K_F2,
	"K_F3": K_F3,
	"K_F4": K_F4,
	"K_F5": K_F5,
	"K_F6": K_F6,
	"K_F7": K_F7,
	"K_F8": K_F8,
	"K_F9": K_F9,
	"K_F10": K_F10,
	"K_F11": K_F11,
	"K_F12": K_F12,
	"K_F13": K_F13,
	"K_F14": K_F14,
	"K_F15": K_F15,
	"K_NUMLOCK": K_NUMLOCK,
	"K_CAPSLOCK": K_CAPSLOCK,
	"K_SCROLLOCK": K_SCROLLOCK,
	"K_RSHIFT": K_RSHIFT,
	"K_LSHIFT": K_LSHIFT,
	"K_RCTRL": K_RCTRL,
	"K_LCTRL": K_LCTRL,
	"K_RALT": K_RALT,
	"K_LALT": K_LALT,
	"K_RMETA": K_RMETA,
	"K_LMETA": K_LMETA,
	"K_LSUPER": K_LSUPER,
	"K_RSUPER": K_RSUPER,
	"K_MODE": K_MODE,
	"K_COMPOSE": K_COMPOSE,
	"K_HELP": K_HELP,
	"K_PRINT": K_PRINT,
	"K_SYSREQ": K_SYSREQ,
	"K_BREAK": K_BREAK,
	"K_MENU": K_MENU,
	"K_POWER": K_POWER,
	"K_EURO": K_EURO,
	"K_UNDO": K_UNDO,
}

var keyNames = map[int]string{}

func init() {
	// Init keyNames from keyCodes
	for name, code := range keyCodes {
		keyNames[code] = name
	}
}
