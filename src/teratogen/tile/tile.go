// tile.go
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

package tile

import (
	"image"
	"math"
	"teratogen/num"
)

// http://www-cs-students.stanford.edu/~amitp/Articles/HexLOS.html

// HexDist returns the hexagonal distance between two points.
func HexDist(p1, p2 image.Point) int {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y

	if math.Signbit(float64(dx)) == math.Signbit(float64(dy)) {
		return int(math.Max(math.Abs(float64(dx)), math.Abs(float64(dy))))
	}
	return int(math.Abs(float64(dx)) + math.Abs(float64(dy)))
}

var HexDirs = []image.Point{{-1, -1}, {0, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 0}}

func HexCircumference(radius int) int {
	if radius == 0 {
		return 1
	}
	return radius * 6
}

// HexCirclePoint returns a point along the edge of a radius sized hexagon
// tile "circle" specified by windingIndex. The HexCircumference(radius)
// consecutive clockwise points on the circle are denoted by consecutive
// windingIndex values.
func HexCirclePoint(radius int, windingIndex int) image.Point {
	if radius == 0 {
		return image.Pt(0, 0)
	}

	sector := num.AbsMod(windingIndex, HexCircumference(radius)) / radius
	offset := num.AbsMod(windingIndex, radius)
	return HexDirs[sector].Mul(radius).Add(HexDirs[(sector+2)%6].Mul(offset))
}
