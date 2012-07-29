/* tile.go

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
	"teratogen/num"
)

// http://www-cs-students.stanford.edu/~amitp/Articles/HexLOS.html

// HexDist returns the hexagonal distance between two points.
func HexDist(p1, p2 image.Point) int {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y

	if num.Isignum(dx) == num.Isignum(dy) {
		return num.Imax(num.Iabs(dx), num.Iabs(dy))
	}
	return num.Iabs(dx) + num.Iabs(dy)
}
