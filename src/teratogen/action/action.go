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

package action

import (
	"image"
	"math/rand"
	"teratogen/entity"
	"teratogen/fov"
	"teratogen/manifold"
	"teratogen/mob"
	"teratogen/tile"
	"teratogen/world"
)

type Action struct {
	World *world.World
}

func New(w *world.World) *Action {
	return &Action{w}
}

type fovvable interface {
	entity.Fov
}

func (a *Action) Footprint(obj entity.Entity, loc manifold.Location) manifold.Footprint {
	if ft, ok := obj.(entity.Footprint); ok {
		// Big entities with a complex footprint.
		return a.World.Manifold.MakeFootprint(ft.Footprint(), loc)
	}
	// Entities with a simple one-cell footprint.
	return manifold.Footprint{image.Pt(0, 0): loc}
}

func (a *Action) AttackMove(obj entity.Entity, vec image.Point) {
	newLoc := a.World.Manifold.Offset(a.Loc(obj), vec)
	footprint := a.Footprint(obj, newLoc)

	for _, loc := range footprint {
		for _, oe := range a.World.Spatial.At(loc) {
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
	a.World.Spatial.Remove(target)
}

func (a *Action) Loc(obj entity.Entity) manifold.Location {
	return a.World.Spatial.Loc(obj)
}

func (a *Action) Fits(obj entity.Entity, loc manifold.Location) bool {
	// TODO: handle footprint stuff.
	if a.World.Terrain(loc).BlocksMove() {
		return false
	}
	for _, oe := range a.World.Spatial.At(loc) {
		if b := oe.Entity.(entity.BlockMove); oe.Entity != obj && b != nil && b.BlocksMove() {
			return false
		}
	}
	return true
}

func (a *Action) Move(obj entity.Entity, vec image.Point) {
	newLoc := a.World.Manifold.Offset(a.Loc(obj), vec)

	if a.Fits(obj, newLoc) {
		if f, ok := obj.(entity.Fov); ok {
			f.MoveFovOrigin(vec)
		}
		a.Place(obj, newLoc)
	}
}

func (a *Action) Place(obj entity.Entity, loc manifold.Location) {
	if a.World.Spatial.Contains(obj) {
		a.World.Spatial.Remove(obj)
	}
	a.World.Spatial.Add(obj, loc)

	if f, ok := obj.(fovvable); ok {
		a.DoFov(f)
	}
	if !a.World.IsAlive(obj) {
		panic("Placed obj not shown alive")
	}
}

func (a *Action) DoFov(obj entity.Entity) {
	// TODO: Parametrisable radius
	const radius = 12
	if f, ok := obj.(fovvable); ok {
		fv := fov.New(
			func(loc manifold.Location) bool { return a.World.Terrain(loc).BlocksSight() },
			func(pt image.Point, loc manifold.Location) { f.MarkFov(pt, loc) },
			a.World.Manifold)
		fv.Run(a.Loc(obj), radius)
	}
}

func (a *Action) RunAI() {
	for actor := a.World.NextActor(); actor != nil; actor = a.World.NextActor() {
		if actor == a.World.Player {
			continue
		}
		a.AttackMove(actor, tile.HexDirs[rand.Intn(6)])
	}
}

func (a *Action) EndTurn() {
	a.RunAI()
	a.World.EndTurn()
}

func (a *Action) IsGameOver() bool {
	obj, _ := a.World.Player.(*mob.PC)
	return !a.World.IsAlive(obj)
}
