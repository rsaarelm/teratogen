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
	"teratogen/gfx"
	"teratogen/manifold"
	"teratogen/mob"
	"teratogen/world"
)

type Mapgen struct {
	world   *world.World
	chart   manifold.Chart
	openSet map[manifold.Location]bool
}

func New(w *world.World) *Mapgen {
	return &Mapgen{world: w}
}

func (m *Mapgen) TestMap(start manifold.Location) {
	m.init(start)
	for y := -17; y < 17; y++ {
		for x := -17; x < 17; x++ {
			m.setTerrain(image.Pt(x, y), world.WallTerrain)
		}
	}

	bounds := image.Rect(-16, -16, 16, 16)
	m.bspRooms(bounds)
	m.extraDoors(bounds)

	m.spawn(m.world.Player, start)

	for i := 0; i < 32; i++ {
		spawnLoc := m.randomLoc()
		spawnMob := mob.New(m.world, &mob.Spec{gfx.ImageSpec{"assets/chars.png", image.Rect(8, 0, 16, 8)}})
		m.spawn(spawnMob, spawnLoc)
	}

	m.world.SetTerrain(m.randomLoc(), world.StairTerrain)
}

func (m *Mapgen) init(start manifold.Location) {
	m.chart = simpleChart(start)
	m.openSet = map[manifold.Location]bool{}
}

func (m *Mapgen) setOpen(loc manifold.Location, isOpen bool) {
	if isOpen {
		m.openSet[loc] = true
	} else {
		delete(m.openSet, loc)
	}
}

func (m *Mapgen) randomLoc() (loc manifold.Location) {
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

func (m *Mapgen) spawn(obj entity.Entity, loc manifold.Location) error {
	if !m.world.Fits(obj, loc) {
		return errors.New("Spawn won't fit")
	}

	m.world.Spatial.Add(obj, loc)

	// Remove the points from the open set if the entity blocks movement.
	if b, ok := obj.(entity.BlockMove); ok && b.BlocksMove() {
		for _, footLoc := range m.world.Spatial.EntityFootprint(obj, loc) {
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

// simpleChart is a chart that pays no attention to portals in the manifold.
type simpleChart manifold.Location

func (s simpleChart) At(pt image.Point) manifold.Location {
	return manifold.Location(s).Add(pt)
}
