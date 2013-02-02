// mapgen.go
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

// Package mapgen defines the system for generating new game levels and
// populating them with entities.
package mapgen

import (
	"errors"
	"image"
	"math/rand"
	"teratogen/entity"
	"teratogen/space"
	"teratogen/world"
)

type Mapgen struct {
	world   *world.World
	chart   space.Chart
	openSet map[space.Location]bool
}

func New(w *world.World) *Mapgen {
	return &Mapgen{world: w}
}

func (m *Mapgen) TestMap(start space.Location, depth int) (entry, exit space.Location) {
	ok := false
	// Try several times until you get one where all slots are filled.
	for i := 0; i < 64; i++ {
		entry, exit, ok = m.tryBuild(start, depth)
		if ok {
			break
		}
	}
	if !ok {
		panic("Couldn't generate map")
	}
	return
}

func (m *Mapgen) tryBuild(start space.Location, depth int) (entry, exit space.Location, ok bool) {
	cg := NewChunkGraph()

	// Set boundaries
	edgeChunk := Chunks()[0] // Solid block
	for i := -5; i <= 5; i++ {
		cg.PlaceDummyChunk(image.Pt(i, -5), edgeChunk)
		cg.PlaceDummyChunk(image.Pt(i, 5), edgeChunk)
		cg.PlaceDummyChunk(image.Pt(-5, i), edgeChunk)
		cg.PlaceDummyChunk(image.Pt(5, i), edgeChunk)
	}

	// Entrance vault
	cg.PlaceChunk(image.Pt(0, 0), Chunks()[1])

	for {
		edge := cg.OpenSlots()
		if len(edge) == 0 {
			break
		}

		slot := edge[rand.Intn(len(edge))]
		chunks := cg.FittingChunks(slot, Chunks())

		if len(chunks) == 0 {
			ok = false
			return
		}
		cg.PlaceChunk(slot, chunks[rand.Intn(len(chunks))])
	}

	cg.Chunks().Place(m.world, start)
	// TODO place exit
	return start, space.Location{}, true
}

func (m *Mapgen) init(start space.Location) {
	m.chart = simpleChart(start)
	m.openSet = map[space.Location]bool{}
}

func (m *Mapgen) setOpen(loc space.Location, isOpen bool) {
	if isOpen {
		m.openSet[loc] = true
	} else {
		delete(m.openSet, loc)
	}
}

func (m *Mapgen) randomLoc() (loc space.Location) {
	// XXX: O(n) time.
	n := rand.Intn(len(m.openSet))
	for k, _ := range m.openSet {
		loc = k
		n--
		if n < 0 {
			break
		}
	}
	return
}

func (m *Mapgen) spawn(obj entity.Entity, loc space.Location) error {
	if !m.world.Fits(obj, loc) {
		return errors.New("Spawn won't fit")
	}

	m.world.Place(obj, loc)

	// Remove the points from the open set if the entity blocks movement.
	if b, ok := obj.(entity.BlockMove); ok && b.BlocksMove() {
		for _, footLoc := range m.world.Manifold.FootprintFor(obj, loc) {
			m.setOpen(footLoc, false)
		}
	}
	return nil
}

func (m *Mapgen) terrain(pt image.Point) world.TerrainData {
	return m.world.Terrain(m.chart.At(pt))
}

func (m *Mapgen) setTerrain(pt image.Point, t world.Terrain) {
	m.world.SetTerrain(m.chart.At(pt), t)
}

func (m *Mapgen) checkSurroundings(loc space.Location, mustBeOpen, mustBeClosed []image.Point) bool {
	for _, offset := range mustBeOpen {
		if _, ok := m.openSet[m.world.Manifold.Offset(loc, offset)]; !ok {
			return false
		}
	}

	for _, offset := range mustBeClosed {
		if _, ok := m.openSet[m.world.Manifold.Offset(loc, offset)]; ok {
			return false
		}
	}

	return true
}

func (m *Mapgen) isExitEnclosure(open space.Location) bool {
	// Look for a suitable enclosure site in the positive x direction from the
	// open location.

	// Don't care if the actual site is open or blocked, it will be replaced
	// with a portal to the next level anyway.
	return m.checkSurroundings(open,
		[]image.Point{{0, 0}},
		[]image.Point{{1, -1}, {2, 0}, {2, 1}, {1, 1}})
}

func (m *Mapgen) isEntryEnclosure(open space.Location) bool {
	return m.checkSurroundings(open,
		[]image.Point{{0, 0}},
		[]image.Point{{-2, -1}, {-2, 0}, {-1, 1}, {-1, -1}})
}

// simpleChart is a chart that pays no attention to portals in the manifold.
type simpleChart space.Location

func (s simpleChart) At(pt image.Point) space.Location {
	return space.Location(s).Add(pt)
}
