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
	"teratogen/entity"
	"teratogen/fov"
	"teratogen/manifold"
	"teratogen/world"
)

type Action struct {
	World *world.World
}

func New(w *world.World) *Action {
	return &Action{w}
}

type fovvable interface {
	entity.Located
	entity.Fov
}

func (a *Action) Move(obj entity.Pos, vec image.Point) {
	newLoc := a.World.Manifold.Offset(obj.Loc(), vec)

	if obj.Fits(newLoc) {
		if f, ok := obj.(entity.Fov); ok {
			f.MoveFovOrigin(vec)
		}
		a.Place(obj, newLoc)
	}
}

func (a *Action) Place(obj entity.Pos, loc manifold.Location) {
	obj.Place(loc)

	if f, ok := obj.(fovvable); ok {
		a.DoFov(f)
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
		fv.Run(f.Loc(), radius)
	}
}
