/* fov.go

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

package fov

import (
	"fmt"
	"image"
	"math"
	"teratogen/space"
	"teratogen/tile"
)

type Fov struct {
	blocksSight func(space.Location) bool
	markSeen    func(image.Point, space.Location)
	spc         *space.Space
}

func New(blocksSightFn func(space.Location) bool,
	markSeenFn func(image.Point, space.Location),
	spc *space.Space) *Fov {
	return &Fov{blocksSightFn, markSeenFn, spc}
}

// Run runs a field-of-view computation up to radius distance from the given
// origin, and calls the MarkSeen callback for all locations it finds visible.
func (f *Fov) Run(origin space.Location, radius int) {
	f.markSeen(image.Pt(0, 0), origin)
	f.process(origin, radius, angle{0, 1}, angle{6, 1})
}

func (f *Fov) process(origin space.Location, radius int, begin, end angle) {
	if begin.radius > radius {
		return
	}

	group := f.group(origin, begin.point())
	for a := begin; a.isBelow(end); a = a.next() {
		pt := a.point()
		if f.group(origin, pt) != group {
			// The type of terrain changed, recurse a deeper process with
			// current arc and start a new arc.
			if !group.blocksSight {
				f.process(origin.Beyond(group.portal), radius, begin.above(), a.above())
			}
			f.process(origin, radius, a, end)
			return
		}
		f.markSeen(pt, f.spc.Offset(origin, pt))
	}
	// Recurse after finishing the whole arc.
	if !group.blocksSight {
		f.process(origin.Beyond(group.portal), radius, begin.above(), end.above())
	}
}

// group is used to define contiguous sets of cells along the fov outer radius
// that can be handled as a single unit. These cells must have an identical
// portal and an identical opaqueness.
type group struct {
	blocksSight bool
	portal      space.Portal
}

func (f *Fov) group(origin space.Location, offset image.Point) group {
	rawLoc := origin.Add(offset)
	return group{f.blocksSight(f.spc.Traverse(rawLoc)), f.spc.Portal(rawLoc)}
}

type angle struct {
	pos    float64
	radius int
}

func (a angle) windingIndex() int {
	return int(math.Floor(a.pos + 0.5))
}

func (a angle) endIndex() int {
	return int(math.Ceil(a.pos + 0.5))
}

func (a angle) isBelow(end angle) bool {
	return a.windingIndex() < end.endIndex()
}

func (a angle) next() (a2 angle) {
	a2 = a
	a2.pos += 0.5
	a2.pos = math.Floor(a2.pos)
	a2.pos += 0.5
	return
}

func (a angle) above() angle {
	return angle{a.pos * float64(a.radius+1) / float64(a.radius), a.radius + 1}
}

func (a angle) point() image.Point {
	return tile.HexCirclePoint(a.radius, a.windingIndex())
}

func (a angle) String() string {
	return fmt.Sprintf("%.2f along %d", a.pos, a.radius)
}
