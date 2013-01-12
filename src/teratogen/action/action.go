// action.go
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

// Package action defines the system for performing complex operations on the
// game world, such as game character behavior.
package action

import (
	"image"
	"math/rand"
	"teratogen/entity"
	"teratogen/fov"
	"teratogen/mapgen"
	"teratogen/mob"
	"teratogen/space"
	"teratogen/tile"
	"teratogen/world"
)

type Action struct {
	world  *world.World
	mapgen *mapgen.Mapgen
}

func New(w *world.World, m *mapgen.Mapgen) *Action {
	return &Action{world: w, mapgen: m}
}

type fovvable interface {
	entity.Fov
}

func (a *Action) Footprint(obj entity.Entity, loc space.Location) space.Footprint {
	if ft, ok := obj.(entity.Footprint); ok {
		// Big entities with a complex footprint.
		return a.world.Manifold.MakeFootprint(ft.Footprint(), loc)
	}
	// Entities with a simple one-cell footprint.
	return space.Footprint{image.Pt(0, 0): loc}
}

func (a *Action) AttackMove(obj entity.Entity, vec image.Point) {
	newLoc := a.world.Manifold.Offset(a.Loc(obj), vec)
	footprint := a.Footprint(obj, newLoc)

	for _, loc := range footprint {
		for _, oe := range a.world.Spatial.At(loc) {
			hit := oe.Entity
			if hit == obj {
				// Ignore self-intersect
				continue
			}
			if a.EnemyOf(obj, hit) {
				a.Attack(obj, hit)
				return
			}
		}
	}
	a.Move(obj, vec)
}

func (a *Action) EnemyOf(obj1, obj2 entity.Entity) bool {
	// TODO better
	return obj1 != obj2
}

func (a *Action) Attack(attacker, target entity.Entity) {
	// TODO better
	// Just straight up kill the target.
	a.world.Spatial.Remove(target)
}

func (a *Action) Loc(obj entity.Entity) space.Location {
	return a.world.Spatial.Loc(obj)
}

func (a *Action) Move(obj entity.Entity, vec image.Point) {
	newLoc := a.world.Manifold.Offset(a.Loc(obj), vec)

	if a.world.Fits(obj, newLoc) {
		if f, ok := obj.(entity.Fov); ok {
			f.MoveFovOrigin(vec)
		}
		a.Place(obj, newLoc)
	}
}

// Place puts an entity in a specific location and performs any necessary
// further actions that should follow after the entity entering the location.
func (a *Action) Place(obj entity.Entity, loc space.Location) {
	if a.world.Spatial.Contains(obj) {
		a.world.Spatial.Remove(obj)
	}
	a.world.Spatial.Add(obj, loc)
	if !a.world.IsAlive(obj) {
		panic("Placed obj not shown alive")
	}

	if f, ok := obj.(fovvable); ok {
		a.DoFov(f)
	}

	for _, footLoc := range a.world.Spatial.EntityFootprint(obj, loc) {
		if obj == a.world.Player && a.world.Terrain(footLoc).Kind == world.StairKind {
			// Player stepping on a stair, go to next level.
			a.NextLevel()
		}
	}
}

func (a *Action) DoFov(obj entity.Entity) {
	// TODO: Parametrisable radius
	const radius = 12
	if f, ok := obj.(fovvable); ok {
		fv := fov.New(
			func(loc space.Location) bool { return a.world.Terrain(loc).BlocksSight() },
			func(pt image.Point, loc space.Location) { f.MarkFov(pt, loc) },
			a.world.Manifold)
		fv.Run(a.Loc(obj), radius)
	}
}

func (a *Action) RunAI() {
	for actor := a.world.NextActor(); actor != nil; actor = a.world.NextActor() {
		if actor == a.world.Player {
			continue
		}
		a.AttackMove(actor, tile.HexDirs[rand.Intn(6)])
	}
}

func (a *Action) EndTurn() {
	a.RunAI()
	a.world.EndTurn()
}

func (a *Action) IsGameOver() bool {
	obj, _ := a.world.Player.(*mob.PC)
	return !a.world.IsAlive(obj)
}

// NextLevel clears out the current level and moves the player to the next one.
func (a *Action) NextLevel() {
	a.world.Floor++
	a.world.Clear()
	if f, ok := a.world.Player.(fovvable); ok {
		f.ClearFov()
	}

	origin := space.Location{0, 0, 1}
	a.mapgen.TestMap(origin)
	a.DoFov(a.world.Player)
}
