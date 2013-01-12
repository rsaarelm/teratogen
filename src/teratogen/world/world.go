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

// Package world defines the structure you want to store when you save the
// game.
package world

import (
	"teratogen/entity"
	"teratogen/gfx"
	"teratogen/space"
	"teratogen/spatial"
)

type World struct {
	Manifold *space.Manifold
	terrain  map[space.Location]Terrain
	Spatial  *spatial.Spatial
	Floor    int
	// Actor queue for the current frame
	actors []entity.Entity
	// Actor queue for the next frame
	nextActors []entity.Entity

	Player interface {
		gfx.Spritable
		entity.Fov
	}
}

func New() (world *World) {
	world = new(World)
	world.Manifold = space.New()
	world.terrain = make(map[space.Location]Terrain)
	world.Spatial = spatial.New(world.Manifold)
	world.actors = []entity.Entity{}
	world.nextActors = []entity.Entity{}
	return
}

func (w *World) Terrain(loc space.Location) TerrainData {
	if t, ok := w.terrain[loc]; ok {
		return terrainTable[t]
	}
	return terrainTable[VoidTerrain]
}

func (w *World) SetTerrain(loc space.Location, t Terrain) {
	w.terrain[loc] = t
}

func (w *World) ClearTerrain() {
	w.terrain = make(map[space.Location]Terrain)
}

func (w *World) Clear() {
	w.ClearTerrain()
	w.Spatial.Clear()
}

func (w *World) Contains(loc space.Location) bool {
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

func (w *World) Fits(obj entity.Entity, loc space.Location) bool {
	for _, footLoc := range w.Spatial.EntityFootprint(obj, loc) {
		if w.Terrain(footLoc).BlocksMove() {
			return false
		}
		for _, oe := range w.Spatial.At(footLoc) {
			if b, ok := oe.Entity.(entity.BlockMove); oe.Entity != obj && ok && b.BlocksMove() {
				return false
			}
		}
	}
	return true
}
