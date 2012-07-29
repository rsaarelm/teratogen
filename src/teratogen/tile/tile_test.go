/* tile_test.go

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

package tile

import (
	"image"
	"testing"
)

func TestHexDist(t *testing.T) {
	if HexDist(image.Pt(10, 10), image.Pt(10, 10)) != 0 {
		t.Fail()
	}

	if HexDist(image.Pt(3, -3), image.Pt(2, -3)) != 1 {
		t.Fail()
	}

	if HexDist(image.Pt(3, -3), image.Pt(4, -2)) != 1 {
		t.Fail()
	}

	if HexDist(image.Pt(3, -3), image.Pt(4, -4)) != 2 {
		t.Fail()
	}
}
