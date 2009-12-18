package event

type Event interface {}

type KeyDown keyEvent

type KeyUp keyEvent

type keyEvent struct {
	KeySym int
	Printable int
	ModifierFlags uint
}

type MouseMove mouseEvent

type MouseDown mouseEvent

type MouseUp mouseEvent

type mouseEvent struct {
	X, Y int
	Dx, Dy int
	ButtonStates uint
	ChangedButton int
}

type Resize struct { Width, Height int }

type Quit struct { }

// Mouse button, modifier key and keysym codes are the same as in SDL, so the
// values from SDL can be passed straight to the event system.

// Mouse buttons
const (
	MOUSE_LEFT = iota + 1
	MOUSE_MIDDLE
	MOUSE_RIGHT
	MOUSE_WHEELUP
	MOUSE_WHEELDOWN
)

// Modifier key flags
const (
	MOD_NONE = 0x0000
	MOD_LSHIFT = 0x0001
	MOD_RSHIFT = 0x0002
	MOD_LCTRL = 0x0040
	MOD_RCTRL = 0x0080
	MOD_LALT = 0x0100
	MOD_RALT = 0x0200
	MOD_LMETA = 0x0400
	MOD_RMETA = 0x0800
	MOD_NUM = 0x1000
	MOD_CAPS = 0x2000
	MOD_MODE = 0x4000
	MOD_CTRL = MOD_LCTRL | MOD_RCTRL
	MOD_SHIFT = MOD_LSHIFT | MOD_RSHIFT
	MOD_ALT = MOD_LALT | MOD_RALT
	MOD_META = MOD_LMETA | MOD_RMETA
)

// Keysyms
const (
	K_BACKSPACE = 8
	K_TAB = 9
	K_CLEAR = 12
	K_RETURN = 13
	K_PAUSE = 19
	K_ESCAPE = 27
	K_SPACE = 32
	K_EXCLAIM = 33
	K_QUOTEDBL = 34
	K_HASH = 35
	K_DOLLAR = 36
	K_AMPERSAND = 38
	K_QUOTE = 39
	K_LEFTPAREN = 40
	K_RIGHTPAREN = 41
	K_ASTERISK = 42
	K_PLUS = 43
	K_COMMA = 44
	K_MINUS = 45
	K_PERIOD = 46
	K_SLASH = 47
	K_0 = 48
	K_1 = 49
	K_2 = 50
	K_3 = 51
	K_4 = 52
	K_5 = 53
	K_6 = 54
	K_7 = 55
	K_8 = 56
	K_9 = 57
	K_COLON = 58
	K_SEMICOLON = 59
	K_LESS = 60
	K_EQUALS = 61
	K_GREATER = 62
	K_QUESTION = 63
	K_AT = 64
	K_LEFTBRACKET = 91
	K_BACKSLASH = 92
	K_RIGHTBRACKET = 93
	K_CARET = 94
	K_UNDERSCORE = 95
	K_BACKQUOTE = 96
	K_A = 97
	K_B = 98
	K_C = 99
	K_D = 100
	K_E = 101
	K_F = 102
	K_G = 103
	K_H = 104
	K_I = 105
	K_J = 106
	K_K = 107
	K_L = 108
	K_M = 109
	K_N = 110
	K_O = 111
	K_P = 112
	K_Q = 113
	K_R = 114
	K_S = 115
	K_T = 116
	K_U = 117
	K_V = 118
	K_W = 119
	K_X = 120
	K_Y = 121
	K_Z = 122
	K_DELETE = 127
	K_KP0 = 256
	K_KP1 = 257
	K_KP2 = 258
	K_KP3 = 259
	K_KP4 = 260
	K_KP5 = 261
	K_KP6 = 262
	K_KP7 = 263
	K_KP8 = 264
	K_KP9 = 265
	K_KP_PERIOD = 266
	K_KP_DIVIDE = 267
	K_KP_MULTIPLY = 268
	K_KP_MINUS = 269
	K_KP_PLUS = 270
	K_KP_ENTER = 271
	K_KP_EQUALS = 272
	K_UP = 273
	K_DOWN = 274
	K_RIGHT = 275
	K_LEFT = 276
	K_INSERT = 277
	K_HOME = 278
	K_END = 279
	K_PAGEUP = 280
	K_PAGEDOWN = 281
	K_F1 = 282
	K_F2 = 283
	K_F3 = 284
	K_F4 = 285
	K_F5 = 286
	K_F6 = 287
	K_F7 = 288
	K_F8 = 289
	K_F9 = 290
	K_F10 = 291
	K_F11 = 292
	K_F12 = 293
	K_F13 = 294
	K_F14 = 295
	K_F15 = 296
	K_NUMLOCK = 300
	K_CAPSLOCK = 301
	K_SCROLLOCK = 302
	K_RSHIFT = 303
	K_LSHIFT = 304
	K_RCTRL = 305
	K_LCTRL = 306
	K_RALT = 307
	K_LALT = 308
	K_RMETA = 309
	K_LMETA = 310
	K_LSUPER = 311
	K_RSUPER = 312
	K_MODE = 313
	K_COMPOSE = 314
	K_HELP = 315
	K_PRINT = 316
	K_SYSREQ = 317
	K_BREAK = 318
	K_MENU = 319
	K_POWER = 320
	K_EURO = 321
	K_UNDO = 322
)
