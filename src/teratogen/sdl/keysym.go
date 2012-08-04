// keysym.go
//
// Copyright (C) 2012 Risto Saarelma
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package sdl

/*
#include <SDL/SDL_keysym.h>
*/
import "C"

import (
	"fmt"
	"strings"
)

// Denotes a key on the keyboard, like 'a' or 'up arrow'.
type KeySym int

// Denotes a nonportable hardware scancode for a keyboard key. Independent of
// keyboard layout.
type Scancode int

type KeyMod int

// SDL key modifiers
const (
	KMOD_NONE   KeyMod = C.KMOD_NONE
	KMOD_LSHIFT        = C.KMOD_LSHIFT
	KMOD_RSHIFT        = C.KMOD_RSHIFT
	KMOD_LCTRL         = C.KMOD_LCTRL
	KMOD_RCTRL         = C.KMOD_RCTRL
	KMOD_LALT          = C.KMOD_LALT
	KMOD_RALT          = C.KMOD_RALT
	KMOD_LMETA         = C.KMOD_LMETA
	KMOD_RMETA         = C.KMOD_RMETA
	KMOD_NUM           = C.KMOD_NUM
	KMOD_CAPS          = C.KMOD_CAPS
	KMOD_MODE          = C.KMOD_MODE

	KMOD_CTRL  = KMOD_LCTRL | KMOD_RCTRL
	KMOD_SHIFT = KMOD_LSHIFT | KMOD_RSHIFT
	KMOD_ALT   = KMOD_LALT | KMOD_RALT
	KMOD_META  = KMOD_RMETA | KMOD_LMETA
)

var modNames = map[KeyMod]string{
	KMOD_LSHIFT: "LSHIFT",
	KMOD_RSHIFT: "RSHIFT",
	KMOD_LCTRL:  "LCTRL",
	KMOD_RCTRL:  "RCTRL",
	KMOD_LALT:   "LALT",
	KMOD_RALT:   "RALT",
	KMOD_LMETA:  "LMETA",
	KMOD_RMETA:  "RMETA",
	KMOD_NUM:    "NUM",
	KMOD_CAPS:   "CAPS",
	KMOD_MODE:   "MODE",
}

func (m KeyMod) String() string {
	mods := make([]string, 0)
	for k, v := range modNames {
		if m&k != 0 {
			mods = append(mods, v)
		}
	}
	return strings.Join(mods, "|")
}

func (sym KeySym) String() string {
	if str, found := symNames[sym]; found {
		return str
	}
	return fmt.Sprintf("KEY_%d", int(sym))
}

func TranslateScancode(code Scancode) KeySym {
	if int(code) >= 0 && int(code) < len(scancodeMap) {
		return scancodeMap[int(code)]
	}
	return K_UNKNOWN
}

// FixedKeySym tries to use the scancode of the key event to get a keyboard
// layout invariant keysym. If this fails, it returns the unchanged keysym.
func (evt KeyEvent) FixedSym() KeySym {
	if t := TranslateScancode(evt.Code); t != K_UNKNOWN {
		return t
	}
	return evt.Sym
}

// SDL keysym constants
const (
	K_UNKNOWN      KeySym = C.SDLK_UNKNOWN
	K_FIRST               = C.SDLK_FIRST
	K_BACKSPACE           = C.SDLK_BACKSPACE
	K_TAB                 = C.SDLK_TAB
	K_CLEAR               = C.SDLK_CLEAR
	K_RETURN              = C.SDLK_RETURN
	K_PAUSE               = C.SDLK_PAUSE
	K_ESCAPE              = C.SDLK_ESCAPE
	K_SPACE               = C.SDLK_SPACE
	K_EXCLAIM             = C.SDLK_EXCLAIM
	K_QUOTEDBL            = C.SDLK_QUOTEDBL
	K_HASH                = C.SDLK_HASH
	K_DOLLAR              = C.SDLK_DOLLAR
	K_AMPERSAND           = C.SDLK_AMPERSAND
	K_QUOTE               = C.SDLK_QUOTE
	K_LEFTPAREN           = C.SDLK_LEFTPAREN
	K_RIGHTPAREN          = C.SDLK_RIGHTPAREN
	K_ASTERISK            = C.SDLK_ASTERISK
	K_PLUS                = C.SDLK_PLUS
	K_COMMA               = C.SDLK_COMMA
	K_MINUS               = C.SDLK_MINUS
	K_PERIOD              = C.SDLK_PERIOD
	K_SLASH               = C.SDLK_SLASH
	K_0                   = C.SDLK_0
	K_1                   = C.SDLK_1
	K_2                   = C.SDLK_2
	K_3                   = C.SDLK_3
	K_4                   = C.SDLK_4
	K_5                   = C.SDLK_5
	K_6                   = C.SDLK_6
	K_7                   = C.SDLK_7
	K_8                   = C.SDLK_8
	K_9                   = C.SDLK_9
	K_COLON               = C.SDLK_COLON
	K_SEMICOLON           = C.SDLK_SEMICOLON
	K_LESS                = C.SDLK_LESS
	K_EQUALS              = C.SDLK_EQUALS
	K_GREATER             = C.SDLK_GREATER
	K_QUESTION            = C.SDLK_QUESTION
	K_AT                  = C.SDLK_AT
	K_LEFTBRACKET         = C.SDLK_LEFTBRACKET
	K_BACKSLASH           = C.SDLK_BACKSLASH
	K_RIGHTBRACKET        = C.SDLK_RIGHTBRACKET
	K_CARET               = C.SDLK_CARET
	K_UNDERSCORE          = C.SDLK_UNDERSCORE
	K_BACKQUOTE           = C.SDLK_BACKQUOTE
	K_a                   = C.SDLK_a
	K_b                   = C.SDLK_b
	K_c                   = C.SDLK_c
	K_d                   = C.SDLK_d
	K_e                   = C.SDLK_e
	K_f                   = C.SDLK_f
	K_g                   = C.SDLK_g
	K_h                   = C.SDLK_h
	K_i                   = C.SDLK_i
	K_j                   = C.SDLK_j
	K_k                   = C.SDLK_k
	K_l                   = C.SDLK_l
	K_m                   = C.SDLK_m
	K_n                   = C.SDLK_n
	K_o                   = C.SDLK_o
	K_p                   = C.SDLK_p
	K_q                   = C.SDLK_q
	K_r                   = C.SDLK_r
	K_s                   = C.SDLK_s
	K_t                   = C.SDLK_t
	K_u                   = C.SDLK_u
	K_v                   = C.SDLK_v
	K_w                   = C.SDLK_w
	K_x                   = C.SDLK_x
	K_y                   = C.SDLK_y
	K_z                   = C.SDLK_z
	K_DELETE              = C.SDLK_DELETE
	K_KP0                 = C.SDLK_KP0
	K_KP1                 = C.SDLK_KP1
	K_KP2                 = C.SDLK_KP2
	K_KP3                 = C.SDLK_KP3
	K_KP4                 = C.SDLK_KP4
	K_KP5                 = C.SDLK_KP5
	K_KP6                 = C.SDLK_KP6
	K_KP7                 = C.SDLK_KP7
	K_KP8                 = C.SDLK_KP8
	K_KP9                 = C.SDLK_KP9
	K_KP_PERIOD           = C.SDLK_KP_PERIOD
	K_KP_DIVIDE           = C.SDLK_KP_DIVIDE
	K_KP_MULTIPLY         = C.SDLK_KP_MULTIPLY
	K_KP_MINUS            = C.SDLK_KP_MINUS
	K_KP_PLUS             = C.SDLK_KP_PLUS
	K_KP_ENTER            = C.SDLK_KP_ENTER
	K_KP_EQUALS           = C.SDLK_KP_EQUALS
	K_UP                  = C.SDLK_UP
	K_DOWN                = C.SDLK_DOWN
	K_RIGHT               = C.SDLK_RIGHT
	K_LEFT                = C.SDLK_LEFT
	K_INSERT              = C.SDLK_INSERT
	K_HOME                = C.SDLK_HOME
	K_END                 = C.SDLK_END
	K_PAGEUP              = C.SDLK_PAGEUP
	K_PAGEDOWN            = C.SDLK_PAGEDOWN
	K_F1                  = C.SDLK_F1
	K_F2                  = C.SDLK_F2
	K_F3                  = C.SDLK_F3
	K_F4                  = C.SDLK_F4
	K_F5                  = C.SDLK_F5
	K_F6                  = C.SDLK_F6
	K_F7                  = C.SDLK_F7
	K_F8                  = C.SDLK_F8
	K_F9                  = C.SDLK_F9
	K_F10                 = C.SDLK_F10
	K_F11                 = C.SDLK_F11
	K_F12                 = C.SDLK_F12
	K_F13                 = C.SDLK_F13
	K_F14                 = C.SDLK_F14
	K_F15                 = C.SDLK_F15
	K_NUMLOCK             = C.SDLK_NUMLOCK
	K_CAPSLOCK            = C.SDLK_CAPSLOCK
	K_SCROLLOCK           = C.SDLK_SCROLLOCK
	K_RSHIFT              = C.SDLK_RSHIFT
	K_LSHIFT              = C.SDLK_LSHIFT
	K_RCTRL               = C.SDLK_RCTRL
	K_LCTRL               = C.SDLK_LCTRL
	K_RALT                = C.SDLK_RALT
	K_LALT                = C.SDLK_LALT
	K_RMETA               = C.SDLK_RMETA
	K_LMETA               = C.SDLK_LMETA
	K_LSUPER              = C.SDLK_LSUPER
	K_RSUPER              = C.SDLK_RSUPER
	K_MODE                = C.SDLK_MODE
	K_COMPOSE             = C.SDLK_COMPOSE
	K_HELP                = C.SDLK_HELP
	K_PRINT               = C.SDLK_PRINT
	K_SYSREQ              = C.SDLK_SYSREQ
	K_BREAK               = C.SDLK_BREAK
	K_MENU                = C.SDLK_MENU
	K_POWER               = C.SDLK_POWER
	K_EURO                = C.SDLK_EURO
	K_UNDO                = C.SDLK_UNDO
	K_LAST                = C.SDLK_LAST
)

var symNames = map[KeySym]string{
	K_BACKSPACE:    "BACKSPACE",
	K_TAB:          "TAB",
	K_CLEAR:        "CLEAR",
	K_RETURN:       "RETURN",
	K_PAUSE:        "PAUSE",
	K_ESCAPE:       "ESCAPE",
	K_SPACE:        "SPACE",
	K_EXCLAIM:      "EXCLAIM",
	K_QUOTEDBL:     "QUOTEDBL",
	K_HASH:         "HASH",
	K_DOLLAR:       "DOLLAR",
	K_AMPERSAND:    "AMPERSAND",
	K_QUOTE:        "QUOTE",
	K_LEFTPAREN:    "LEFTPAREN",
	K_RIGHTPAREN:   "RIGHTPAREN",
	K_ASTERISK:     "ASTERISK",
	K_PLUS:         "PLUS",
	K_COMMA:        "COMMA",
	K_MINUS:        "MINUS",
	K_PERIOD:       "PERIOD",
	K_SLASH:        "SLASH",
	K_0:            "0",
	K_1:            "1",
	K_2:            "2",
	K_3:            "3",
	K_4:            "4",
	K_5:            "5",
	K_6:            "6",
	K_7:            "7",
	K_8:            "8",
	K_9:            "9",
	K_COLON:        "COLON",
	K_SEMICOLON:    "SEMICOLON",
	K_LESS:         "LESS",
	K_EQUALS:       "EQUALS",
	K_GREATER:      "GREATER",
	K_QUESTION:     "QUESTION",
	K_AT:           "AT",
	K_LEFTBRACKET:  "LEFTBRACKET",
	K_BACKSLASH:    "BACKSLASH",
	K_RIGHTBRACKET: "RIGHTBRACKET",
	K_CARET:        "CARET",
	K_UNDERSCORE:   "UNDERSCORE",
	K_BACKQUOTE:    "BACKQUOTE",
	K_a:            "A",
	K_b:            "B",
	K_c:            "C",
	K_d:            "D",
	K_e:            "E",
	K_f:            "F",
	K_g:            "G",
	K_h:            "H",
	K_i:            "I",
	K_j:            "J",
	K_k:            "K",
	K_l:            "L",
	K_m:            "M",
	K_n:            "N",
	K_o:            "O",
	K_p:            "P",
	K_q:            "Q",
	K_r:            "R",
	K_s:            "S",
	K_t:            "T",
	K_u:            "U",
	K_v:            "V",
	K_w:            "W",
	K_x:            "X",
	K_y:            "Y",
	K_z:            "Z",
	K_DELETE:       "DELETE",
	K_KP0:          "KP0",
	K_KP1:          "KP1",
	K_KP2:          "KP2",
	K_KP3:          "KP3",
	K_KP4:          "KP4",
	K_KP5:          "KP5",
	K_KP6:          "KP6",
	K_KP7:          "KP7",
	K_KP8:          "KP8",
	K_KP9:          "KP9",
	K_KP_PERIOD:    "KP_PERIOD",
	K_KP_DIVIDE:    "KP_DIVIDE",
	K_KP_MULTIPLY:  "KP_MULTIPLY",
	K_KP_MINUS:     "KP_MINUS",
	K_KP_PLUS:      "KP_PLUS",
	K_KP_ENTER:     "KP_ENTER",
	K_KP_EQUALS:    "KP_EQUALS",
	K_UP:           "UP",
	K_DOWN:         "DOWN",
	K_RIGHT:        "RIGHT",
	K_LEFT:         "LEFT",
	K_INSERT:       "INSERT",
	K_HOME:         "HOME",
	K_END:          "END",
	K_PAGEUP:       "PAGEUP",
	K_PAGEDOWN:     "PAGEDOWN",
	K_F1:           "F1",
	K_F2:           "F2",
	K_F3:           "F3",
	K_F4:           "F4",
	K_F5:           "F5",
	K_F6:           "F6",
	K_F7:           "F7",
	K_F8:           "F8",
	K_F9:           "F9",
	K_F10:          "F10",
	K_F11:          "F11",
	K_F12:          "F12",
	K_F13:          "F13",
	K_F14:          "F14",
	K_F15:          "F15",
	K_NUMLOCK:      "NUMLOCK",
	K_CAPSLOCK:     "CAPSLOCK",
	K_SCROLLOCK:    "SCROLLOCK",
	K_RSHIFT:       "RSHIFT",
	K_LSHIFT:       "LSHIFT",
	K_RCTRL:        "RCTRL",
	K_LCTRL:        "LCTRL",
	K_RALT:         "RALT",
	K_LALT:         "LALT",
	K_RMETA:        "RMETA",
	K_LMETA:        "LMETA",
	K_LSUPER:       "LSUPER",
	K_RSUPER:       "RSUPER",
	K_MODE:         "MODE",
	K_COMPOSE:      "COMPOSE",
	K_HELP:         "HELP",
	K_PRINT:        "PRINT",
	K_SYSREQ:       "SYSREQ",
	K_BREAK:        "BREAK",
	K_MENU:         "MENU",
	K_POWER:        "POWER",
	K_EURO:         "EURO",
	K_UNDO:         "UNDO",
}
