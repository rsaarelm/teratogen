// mob.go
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

package mob

import (
	"image"
	"teratogen/entity"
	"teratogen/gfx"
	"teratogen/manifold"
	"teratogen/world"
	"time"
)

type Mob struct {
	icon   gfx.ImageSpec
	loc    manifold.Location
	world  *world.World
	placed bool
}

type Spec struct {
	Icon gfx.ImageSpec
}

func New(w *world.World, spec *Spec) (result *Mob) {
	result = new(Mob)
	result.world = w

	result.icon = spec.Icon

	return
}

func (m *Mob) DrawLayer() int {
	return 1000
}

func (m *Mob) Icon() gfx.ImageSpec {
	return m.icon
}

func (m *Mob) Sprite(context gfx.Context, offset image.Point) gfx.Sprite {
	// Add bob animation to animate creatures.
	bob := image.Pt(0, 0)
	if time.Now().Nanosecond()%500e6 < 250e6 {
		bob = image.Pt(0, -1)
	}

	return gfx.Sprite{
		Layer:    1000,
		Drawable: context.GetDrawable(m.icon),
		Offset:   offset.Add(bob)}
}

func (m *Mob) Loc() manifold.Location {
	return m.loc
}

func (m *Mob) Place(loc manifold.Location) {
	if m.placed {
		m.Remove()
	}
	m.loc = loc
	m.world.Spatial.Add(m, m.loc)
	m.placed = true
}

func (m *Mob) Remove() {
	if m.placed {
		m.world.Spatial.Remove(m)
		m.placed = false
	}
}

func (m *Mob) Fits(loc manifold.Location) bool {
	if m.world.Terrain(loc).BlocksMove() {
		return false
	}
	for _, oe := range m.world.Spatial.At(loc) {
		if b := oe.Entity.(entity.BlockMove); oe.Entity != m && b != nil && b.BlocksMove() {
			return false
		}
	}
	return true
}

func (m *Mob) BlocksMove() bool {
	return true
}
