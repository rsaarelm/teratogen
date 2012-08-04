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

package manifold

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
// footprints. It is an ordered sequence of unit distance steps that form the
// entire footprint and always extend from either point (0, 0) or steps made
// earlier in the sequence. The expansion of a multi-tile entity's footprint
// must be made explicit in this way to correctly account for situations where
// the entity has partially moved through a portal.
type FootprintTemplate struct {
	steps []footprintStep
}

func NewTemplate() *FootprintTemplate {
	return &FootprintTemplate{make([]footprintStep, 0)}
}

// AddStep adds a construction step to the footprint template. The new point's
// parent point must be either (0, 0) or a point added by a previous step, and
// the new point must be at distance 1 from the parent point. This method is
// for efficient construction of footprint templates, so checking the
// requirements is the responsibility of the caller.
func (ft *FootprintTemplate) AddStep(parent, pos image.Point) {
	ft.steps = append(ft.steps, footprintStep{parent, pos})
}

func (ft *FootprintTemplate) Validate() error {
	validParents := map[image.Point]bool{image.Pt(0, 0): true}
	for _, e := range ft.steps {
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
func MakeTemplate(shape []image.Point) (result *FootprintTemplate, err error) {
	result = NewTemplate()

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
					result.AddStep(parent, pt)
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

	return
}

// Footprint describes the locations which a multi-cell object occupies on a
// map. It maps the points in the body of the object to locations. Since
// portals can make the manifold non-euclidean, the mapping is not trivial and
// must be computed and stored explicitly.
type Footprint map[image.Point]Location

func (t *Manifold) MakeFootprint(template *FootprintTemplate, loc Location) (result Footprint) {
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
