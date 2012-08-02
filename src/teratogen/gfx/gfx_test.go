/* gfx_test.go

   Copyright (C) 2012 Risto Saarelma

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package gfx

import (
	"image/color"
	"testing"
)

func badColor(t *testing.T, str string) {
	if _, ok := ParseColor(str); ok {
		t.Errorf("Parsed a non-color")
	}
}

func goodColor(t *testing.T, str string, expected color.Color) {
	if col, ok := ParseColor(str); !ok {
		t.Errorf("Couldn't parse valid color %s", str)
	} else {
		if col != expected {
			t.Errorf("Expected %s for '%s', parse returned %s", expected, str, col)
		}
	}
}

func TestParseColor(t *testing.T) {
	goodColor(t, "beige", Beige)
	goodColor(t, "BeiGE", Beige)
	badColor(t, "xbeige")
	badColor(t, "SquamousAndRugoseWithAHintOfLavender")
	goodColor(t, "#89abCD", color.RGBA{0x89, 0xab, 0xcd, 0xff})
	goodColor(t, "#8aC", color.RGBA{0x88, 0xaa, 0xcc, 0xff})
	badColor(t, "#12")
	badColor(t, "#1234")
	badColor(t, "#1234567")
	badColor(t, "0#123")
}
