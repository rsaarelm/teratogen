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
	"teratogen/num"
	"teratogen/world"
	"time"
	"unsafe"
)

type Mob struct {
	icon   gfx.ImageSpec
	loc    manifold.Location
	world  *world.World
	placed bool
}

type PC struct {
	Mob
	Fov
}

func NewPC(w *world.World, spec *Spec) (result *PC) {
	result = new(PC)
	result.Mob.Init(w, spec)
	result.Fov.Init()
	return
}

type Spec struct {
	Icon gfx.ImageSpec
}

func New(w *world.World, spec *Spec) (result *Mob) {
	result = new(Mob)
	result.Init(w, spec)
	return
}

func (m *Mob) Init(w *world.World, spec *Spec) {
	m.world = w
	m.icon = spec.Icon
	w.AddActor(m)
}

func (m *Mob) DrawLayer() int {
	return 1000
}

func (m *Mob) Icon() gfx.ImageSpec {
	return m.icon
}

// bob returns the motion offset for the idle animation of the mob's sprite.
func (m *Mob) bob() image.Point {
	t := time.Now().UnixNano()

	// Give different mobs persistent random phases to their bob with noise
	// generated from the mob's pointer value.
	t += int64(1e9 * num.Noise(int(uintptr(unsafe.Pointer(m)))))

	if t%500e6 < 250e6 {
		return image.Pt(0, -1)
	}

	return image.Pt(0, 0)
}

func (m *Mob) Sprite(context gfx.Context, offset image.Point) gfx.Sprite {
	return gfx.Sprite{
		Layer:    1000,
		Drawable: context.GetDrawable(m.icon),
		Offset:   offset.Add(m.bob())}
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

func (m *Mob) Move(vec image.Point) bool {
	newLoc := m.world.Manifold.Offset(m.Loc(), vec)

	if m.Fits(newLoc) {
		m.Place(newLoc)
		return true
	}
	return false
}
