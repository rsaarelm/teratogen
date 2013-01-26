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
	"teratogen/display/fx"
	"teratogen/entity"
	"teratogen/fov"
	"teratogen/mapgen"
	"teratogen/query"
	"teratogen/space"
	"teratogen/tile"
	"teratogen/world"
)

type Action struct {
	world  *world.World
	mapgen *mapgen.Mapgen
	query  *query.Query
	fx     *fx.Fx
}

func New(w *world.World, m *mapgen.Mapgen, q *query.Query, f *fx.Fx) *Action {
	return &Action{world: w, mapgen: m, query: q, fx: f}
}

func (a *Action) AttackMove(obj entity.Entity, vec image.Point) {
	newLoc := a.world.Manifold.Offset(a.query.Loc(obj), vec)
	footprint := a.query.Footprint(obj, newLoc)

	for _, loc := range footprint {
		for _, oe := range a.world.Spatial.At(loc) {
			hit := oe.Entity
			if hit == obj {
				// Ignore self-intersect
				continue
			}
			if a.query.EnemyOf(obj, hit) {
				a.Attack(obj, hit)
				return
			}
		}
	}
	a.Move(obj, vec)
}

func (a *Action) Attack(attacker, target entity.Entity) {
	a.Damage(target, 1)
}

func (a *Action) Damage(target entity.Entity, amount int) {
	if mob, ok := target.(entity.Stats); ok {
		mob.Damage(amount)
		if mob.Health() <= 0 {
			// Target died.
			// Extra logic hooks here.
			a.world.Spatial.Remove(target)
		}
	}
}

func (a *Action) Move(obj entity.Entity, vec image.Point) {
	newLoc := a.world.Manifold.Offset(a.query.Loc(obj), vec)

	if a.world.Fits(obj, newLoc) {
		if f, ok := obj.(entity.Fov); ok {
			f.MoveFovOrigin(vec)
		}
		a.Place(obj, newLoc)
	}
}

func (a *Action) Shoot(obj entity.Entity, vec image.Point) {
	// TODO: Determine how far it goes, what it hits...
	a.fx.Beam(a.query.Loc(obj), vec, 6, fx.GunBeam)
}

// Place puts an entity in a specific location and performs any necessary
// further actions that should follow after the entity entering the location.
func (a *Action) Place(obj entity.Entity, loc space.Location) {
	a.world.Place(obj, loc)
	if !a.world.IsAlive(obj) {
		panic("Placed obj not shown alive")
	}

	if f, ok := obj.(entity.Fov); ok {
		a.DoFov(f)
	}

	for _, footLoc := range a.query.Footprint(obj, loc) {
		if obj == a.world.Player && a.world.Terrain(footLoc).Kind == world.StairKind {
			// Player stepping on a stair, go to next level.
			a.NextLevel()
		}
	}
}

func (a *Action) DoFov(obj entity.Entity) {
	// TODO: Parametrisable radius
	const radius = 12
	if f, ok := obj.(entity.Fov); ok {
		fv := fov.New(
			func(loc space.Location) bool { return a.world.Terrain(loc).BlocksSight() },
			func(pt image.Point, loc space.Location) { f.MarkFov(pt, loc) },
			a.world.Manifold)
		fv.Run(a.query.Loc(obj), radius)
	}
}

func (a *Action) RunAI() {
	for actor := a.world.NextActor(); actor != nil; actor = a.world.NextActor() {
		if actor == a.world.Player {
			continue
		}

		moveDir := tile.HexDirs[rand.Intn(6)]
		if enemy, found := a.query.ClosestEnemy(actor); found {
			moveDir = tile.HexVecToDir(enemy.Offset)
		}
		a.AttackMove(actor, moveDir)
	}
}

func (a *Action) EndTurn() {
	a.RunAI()
	a.world.EndTurn()
}

// NextLevel clears out the current level and moves the player to the next one.
func (a *Action) NextLevel() {
	a.world.Floor++
	a.world.Clear()
	if f, ok := a.world.Player.(entity.Fov); ok {
		f.ClearFov()
	}

	origin := space.Location{0, 0, 1}
	a.mapgen.TestMap(origin)
	a.DoFov(a.world.Player)
}
