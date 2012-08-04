// keycode.go
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

// +build !windows !darwin

package sdl

// Hardware keycodes for Linux, also the default for unknown OSes.

var scancodeMap = []KeySym{
	K_UNKNOWN,
	K_UNKNOWN,
	K_UNKNOWN,
	K_UNKNOWN,
	K_UNKNOWN,
	K_UNKNOWN,
	K_UNKNOWN,
	K_UNKNOWN,
	K_UNKNOWN,
	K_ESCAPE,
	K_1,
	K_2,
	K_3,
	K_4,
	K_5,
	K_6,
	K_7,
	K_8,
	K_9,
	K_0,
	K_MINUS,
	K_EQUALS,
	K_BACKSPACE,
	K_TAB,
	K_q,
	K_w,
	K_e,
	K_r,
	K_t,
	K_y,
	K_u,
	K_i,
	K_o,
	K_p,
	K_LEFTBRACKET,
	K_RIGHTBRACKET,
	K_RETURN,
	K_LCTRL,
	K_a,
	K_s,
	K_d,
	K_f,
	K_g,
	K_h,
	K_j,
	K_k,
	K_l,
	K_SEMICOLON,
	K_QUOTE,
	K_BACKQUOTE,
	K_LSHIFT,
	K_BACKSLASH,
	K_z,
	K_x,
	K_c,
	K_v,
	K_b,
	K_n,
	K_m,
	K_COMMA,
	K_PERIOD,
	K_SLASH,
	K_RSHIFT,
	K_KP_MULTIPLY,
	K_LALT,
	K_SPACE,
	K_CAPSLOCK,
	K_F1,
	K_F2,
	K_F3,
	K_F4,
	K_F5,
	K_F6,
	K_F7,
	K_F8,
	K_F9,
	K_F10,
	K_NUMLOCK,
	K_SCROLLOCK,
	K_KP7,
	K_KP8,
	K_KP9,
	K_KP_MINUS,
	K_KP4,
	K_KP5,
	K_KP6,
	K_KP_PLUS,
	K_KP1,
	K_KP2,
	K_KP3,
	K_KP0,
	K_KP_PERIOD,
	K_UNKNOWN,
	K_UNKNOWN,
	K_LESS,
	K_F11,
	K_F12,
}
