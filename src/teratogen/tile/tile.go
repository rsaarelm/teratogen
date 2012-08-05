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

// HexWallType returns the wall tile offset base on the binary mask built from
// its six neighboring walls. The mask starts at the neighbor at (-1, -1) at
// bit 0, and proceeds to the rest of the neighbors clockwise from there.
//
// A bit 1 indicates a wall-type tile at that neighbor position. The result
// value is between 0 and 3:
//
//     0: Pillar (o)
//     1: x-axis wall (\)
//     2: y-axis wall (/)
//     3: xy-diagonal wall (|)
func HexWallType(edgeMask int) int {
	// Table made by going through the 64 combinations by hand and taking a
	// guess at the best-looking central wall-piece for each. Re-tweak as
	// needed.
	//
	//     00  .    01  #    02  .    03  #    04  .    05  #    06  .    07  #
	//       .   .    .   .    .   #    .   #    .   .    .   .    .   #    .   #
	//         *        *        *        *        *        *        *        *
	//       .   .    .   .    .   .    .   .    .   #    .   #    .   #    .   #
	//         .        .        .        .        .        .        .        .
	//
	//     08  .    09  #    10  .    11  #    12  .    13  #    14  .    15  #
	//       .   .    .   .    .   #    .   #    .   .    .   .    .   #    .   #
	//         *        *        *        *        *        *        *        *
	//       .   .    .   .    .   .    .   .    .   #    .   #    .   #    .   #
	//         #        #        #        #        #        #        #        #
	//
	//     16  .    17  #    18  .    19  #    20  .    21  #    22  .    23  #
	//       .   .    .   .    .   #    .   #    .   .    .   .    .   #    .   #
	//         *        *        *        *        *        *        *        *
	//       #   .    #   .    #   .    #   .    #   #    #   #    #   #    #   #
	//         .        .        .        .        .        .        .        .
	//
	//     24  .    25  #    26  .    27  #    28  .    29  #    30  .    31  #
	//       .   .    .   .    .   #    .   #    .   .    .   .    .   #    .   #
	//         *        *        *        *        *        *        *        *
	//       #   .    #   .    #   .    #   .    #   #    #   #    #   #    #   #
	//         #        #        #        #        #        #        #        #
	//
	//     32  .    33  #    34  .    35  #    36  .    37  #    38  .    39  #
	//       #   .    #   .    #   #    #   #    #   .    #   .    #   #    #   #
	//         *        *        *        *        *        *        *        *
	//       .   .    .   .    .   .    .   .    .   #    .   #    .   #    .   #
	//         .        .        .        .        .        .        .        .
	//
	//     40  .    41  #    42  .    43  #    44  .    45  #    46  .    47  #
	//       #   .    #   .    #   #    #   #    #   .    #   .    #   #    #   #
	//         *        *        *        *        *        *        *        *
	//       .   .    .   .    .   .    .   .    .   #    .   #    .   #    .   #
	//         #        #        #        #        #        #        #        #
	//
	//     48  .    49  #    50  .    51  #    52  .    53  #    54  .    55  #
	//       #   .    #   .    #   #    #   #    #   .    #   .    #   #    #   #
	//         *        *        *        *        *        *        *        *
	//       #   .    #   .    #   .    #   .    #   #    #   #    #   #    #   #
	//         .        .        .        .        .        .        .        .
	//
	//     56  .    57  #    58  .    59  #    60  .    61  #    62  .    63  #
	//       #   .    #   .    #   #    #   #    #   .    #   .    #   #    #   #
	//         *        *        *        *        *        *        *        *
	//       #   .    #   .    #   .    #   .    #   #    #   #    #   #    #   #
	//         #        #        #        #        #        #        #        #
	//

	walls := [64]int{
		0, 0, 2, 2, 1, 1, 0, 0,
		3, 3, 2, 3, 1, 3, 0, 3,
		2, 2, 2, 2, 0, 0, 2, 0,
		2, 3, 2, 0, 0, 0, 2, 2,
		1, 1, 0, 0, 1, 1, 1, 1,
		1, 0, 0, 0, 1, 0, 0, 1,
		0, 0, 2, 2, 1, 0, 0, 0,
		0, 3, 0, 2, 1, 1, 0, 0}
	return walls[edgeMask]
}
