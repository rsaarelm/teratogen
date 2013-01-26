// query.go
//
// Copyright (C) 2013 Risto Saarelma
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

// Package query defines methods for complex queries about the game world
// state.
package query

import (
	"teratogen/entity"
	"teratogen/mob"
	"teratogen/space"
	"teratogen/world"
)

type Query struct {
	world *world.World
}

func New(w *world.World) *Query {
	return &Query{world: w}
}

func (q *Query) Footprint(obj entity.Entity, loc space.Location) space.Footprint {
	return q.world.Manifold.FootprintFor(obj, loc)
}

func (q *Query) EnemyOf(obj1, obj2 entity.Entity) bool {
	// TODO better
	return obj1 != obj2
}

func (q *Query) Loc(obj entity.Entity) space.Location {
	return q.world.Spatial.Loc(obj)
}

func (q *Query) IsGameOver() bool {
	obj, _ := q.world.Player.(*mob.PC)
	return !q.world.IsAlive(obj)
}
