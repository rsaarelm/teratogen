/* footprint.go

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

package space

import (
	"errors"
	"fmt"
	"image"
	"teratogen/tile"
)

type footprintStep struct {
	parent, pos image.Point
}

// FootprintTemplate is a precomputed structure for efficiently generating
// footprints.
type FootprintTemplate struct {
	steps []footprintStep
}

func (it FootprintTemplate) Validate() error {
	validParents := map[image.Point]bool{image.Pt(0, 0): true}
	for _, e := range it.steps {
		if _, ok := validParents[e.parent]; !ok {
			return errors.New(fmt.Sprintf("Unparented node %s", e))
		}
		validParents[e.pos] = true

		if tile.HexDist(e.parent, e.pos) != 1 {
			return errors.New(fmt.Sprintf("Bad parent distance %s", e))
		}
	}
	return nil
}

// MakeTemplate builds an FootprintTemplate from a shape described as a point
// set. The points must be contiguous. The origin point (0, 0) is assumed to
// be included whether or not the shape list contains it, and the rest of the
// points must be contiguous with the origin. Calling MakeTemplate with an
// empty list corresponds to a single-cell footprint that only contains the
// origin point. The connectedness of points is currently determined by hex
// tile distance metric.
func MakeTemplate(shape []image.Point) (result FootprintTemplate, err error) {
	result.steps = make([]footprintStep, 0)

	// Collect points to a map where they can be conveniently deleted from
	// during computation.
	pointSet := make(map[image.Point]bool)
	for _, e := range shape {
		pointSet[e] = true
	}

	parents := []image.Point{image.Pt(0, 0)}
	for len(pointSet) > 0 {
		oldLen := len(pointSet)
		for pt, _ := range pointSet {
			// XXX: Ok to modify the containers during loop?
			for _, parent := range parents {
				if parent == pt {
					// Delete points already in parent set.
					delete(pointSet, pt)
					break
				} else if tile.HexDist(pt, parent) == 1 {
					// XXX: Might want to parameterize the predicate if we
					// ever want to do non-hex geometries?

					// Found a contiguous point.
					result.steps = append(result.steps, footprintStep{parent, pt})
					parents = append(parents, pt)
					delete(pointSet, pt)
					break
				}
			}
		}
		if len(pointSet) == oldLen {
			err = errors.New("Could not form contiguous footprint")
			return
		}
	}

	// XXX: Debug sanity check, remove to make code efficient...
	if err := result.Validate(); err != nil {
		panic(err)
	}

	return
}

// Footprint describes the locations which a multi-cell object occupies on a
// map. It maps the points in the body of the object to locations. Since
// portals can make the space non-euclidean, the mapping is not trivial and
// must be computed and stored explicitly.
type Footprint map[image.Point]Location

func (t *Space) MakeFootprint(template FootprintTemplate, loc Location) (result Footprint) {
	result = map[image.Point]Location{image.Pt(0, 0): loc}

	for _, step := range template.steps {
		parentLoc, ok := result[step.parent]
		if !ok {
			panic("Invalid FootprintTemplate")
		}
		vec := step.pos.Sub(step.parent)
		loc := t.Offset(parentLoc, vec)
		result[step.pos] = loc
	}
	return
}
