// util.go
//
// Copyright (C) 2013 Risto Saarelma
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

package chunk

import (
	"fmt"
	"image"
	"strings"
)

type MapCell rune

const spaceCell = MapCell(' ')

type Dir4 uint8

const (
	north Dir4 = iota
	east
	south
	west
)

func (d Dir4) opposite() Dir4 {
	return (d + 2) % 4
}

func (d Dir4) normalVec() image.Point {
	return []image.Point{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}[d]
}

func (d Dir4) alongVec() image.Point {
	return []image.Point{{1, 0}, {0, 1}, {1, 0}, {0, 1}}[d]
}

type Peg struct {
	offset  image.Point
	facing  Dir4
	pattern string
}

func (p Peg) Add(vec image.Point) Peg {
	return Peg{p.offset.Add(vec), p.facing, p.pattern}
}

func (p Peg) Pattern() []MapCell {
	return []MapCell(p.pattern)
}

// Points returns the geometric points of the Peg, so that the cell at
// peg.Points()[i] is peg.Pattern()[i] for all i.
func (p Peg) Points() []image.Point {
	vec := p.facing.alongVec()
	result := []image.Point{p.offset}
	for i := 1; i < len(p.pattern); i++ {
		result = append(result, result[len(result)-1].Add(vec))
	}
	return result
}

func (p Peg) String() string {
	orient := "|"
	if p.facing == north || p.facing == south {
		orient = "-"
	}
	return fmt.Sprintf("{(%d, %d) %s x %d}",
		p.offset.X, p.offset.Y, orient, len(p.pattern))
}

type mappable interface {
	At(image.Point) (MapCell, bool)
}

func (p Peg) matches(chart mappable) bool {
	pat := p.Pattern()
	for i, pt := range p.Points() {
		if cell, ok := chart.At(pt); ok && cell != pat[i] {
			return false
		}
	}
	return true
}

func (p Peg) covered(chart mappable) bool {
	pts := p.Points()
	// Skip the start and end points, they can be covered even for a Peg
	// considered open. Except if the peg consists only of those points.
	if len(pts) > 2 {
		pts = pts[1 : len(pts)-1]
	}
	for _, pt := range pts {
		if _, ok := chart.At(pt); ok {
			return true
		}
	}
	return false
}

func (p Peg) latches(opposite Peg) bool {
	return opposite.pattern == p.pattern && opposite.facing.opposite() == p.facing
}

type ParseSpec struct {
	PegCells    string
	OverlapCell uint8
}

func (p ParseSpec) isPeg(ch uint8) bool {
	return strings.ContainsRune(p.PegCells, rune(ch))
}

func (p ParseSpec) isOverlap(ch uint8) bool {
	return ch == p.OverlapCell
}

func SplitMaps(lineSeparatedAsciiMaps string) []string {
	result := []string{}
	inBlock := false

	for _, line := range strings.Split(lineSeparatedAsciiMaps, "\n") {
		if strings.TrimSpace(line) == "" {
			inBlock = false
			continue
		}

		if !inBlock {
			result = append(result, "")
			inBlock = true
		}
		result[len(result)-1] += line + "\n"
	}
	return result
}
