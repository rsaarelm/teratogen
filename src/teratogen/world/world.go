// world.go
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

package world

import (
	"image"
	"teratogen/entity"
	"teratogen/gfx"
	"teratogen/manifold"
	"teratogen/mapgen"
	"teratogen/spatial"
)

type World struct {
	Manifold *manifold.Manifold
	terrain  map[manifold.Location]Terrain
	Spatial  *spatial.Spatial
	// Actor queue for the current frame
	actors []entity.Entity
	// Actor queue for the next frame
	nextActors []entity.Entity

	Player interface {
		gfx.Spritable
		entity.Fov
	}
}

type WorldFormer struct {
	world *World
	chart manifold.Chart
}

func (w WorldFormer) At(p image.Point) mapgen.Terrain {
	loc := w.chart.At(p)
	if t, ok := w.world.terrain[loc]; ok {
		switch t {
		case WallTerrain:
			return mapgen.Solid
		case FloorTerrain:
			return mapgen.Open
		case DoorTerrain:
			return mapgen.Doorway
		}
	}

	return mapgen.Solid
}

func (w WorldFormer) Set(p image.Point, t mapgen.Terrain) {
	loc := w.chart.At(p)
	switch t {
	case mapgen.Solid:
		w.world.terrain[loc] = WallTerrain
	case mapgen.Open:
		w.world.terrain[loc] = FloorTerrain
	case mapgen.Doorway:
		w.world.terrain[loc] = DoorTerrain
	}
}

func New() (world *World) {
	world = new(World)
	world.Manifold = manifold.New()
	world.terrain = make(map[manifold.Location]Terrain)
	world.Spatial = spatial.New()
	world.actors = []entity.Entity{}
	world.nextActors = []entity.Entity{}
	return
}

func (w *World) TestMap(origin manifold.Location) {
	bounds := image.Rect(-16, -16, 16, 16)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			w.terrain[origin.Add(image.Pt(x, y))] = WallTerrain
		}
	}
	mapgen.BspRooms(WorldFormer{w, simpleChart(origin)}, bounds.Inset(1))
}

// simpleChart is a chart that pays no attention to portals in the manifold.
type simpleChart manifold.Location

func (s simpleChart) At(pt image.Point) manifold.Location {
	return manifold.Location(s).Add(pt)
}

func (w *World) Terrain(loc manifold.Location) TerrainData {
	if t, ok := w.terrain[loc]; ok {
		return terrainTable[t]
	}
	return terrainTable[VoidTerrain]
}

func (w *World) Contains(loc manifold.Location) bool {
	_, ok := w.terrain[loc]
	return ok
}

func (w *World) AddActor(obj entity.Entity) {
	w.actors = append(w.actors, obj)
}

func (w *World) IsAlive(obj entity.Entity) bool {
	return w.Spatial.Contains(obj)
}

// NextActor returns the next living actor that has still to act this turn.
func (w *World) NextActor() entity.Entity {
	if len(w.actors) == 0 {
		return nil
	}
	actor := w.actors[0]
	w.actors = w.actors[1:]
	if !w.IsAlive(actor) {
		// Drop actors that don't exist in the world.
		return w.NextActor()
	}
	// Otherwise move the actor to the back of the queue and return it.
	w.nextActors = append(w.nextActors, actor)
	return actor
}

// EndTurn refreshes the actor queue for a new game turn.
func (w *World) EndTurn() {
	// Add remaining live actors to the next actor queue.
	for _, a := range w.actors {
		if w.IsAlive(a) {
			w.nextActors = append(w.nextActors, a)
		}
	}
	w.actors = w.nextActors
	w.nextActors = []entity.Entity{}
}
