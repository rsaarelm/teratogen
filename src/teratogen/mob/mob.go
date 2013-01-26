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

// Package mob defines the types for the creatures in Teratogen.
package mob

import (
	"image"
	"teratogen/display/util"
	"teratogen/gfx"
	"teratogen/num"
	"teratogen/space"
	"teratogen/world"
	"time"
	"unsafe"
)

type Mob struct {
	icon      gfx.ImageSpec
	loc       space.Location
	world     *world.World
	placed    bool
	health    int
	maxHealth int
	shield    int
}

type PC struct {
	Mob
	Fov
}

func NewPC(w *world.World, spec Spec) (result *PC) {
	result = new(PC)
	result.Mob.Init(w, spec)
	result.Fov.Init()
	return
}

type Spec struct {
	Icon      gfx.ImageSpec
	MaxHealth int
}

func New(w *world.World, spec Spec) (result *Mob) {
	result = new(Mob)
	result.Init(w, spec)
	return
}

func (m *Mob) Init(w *world.World, spec Spec) {
	m.world = w
	m.icon = spec.Icon
	m.health = spec.MaxHealth
	m.maxHealth = spec.MaxHealth
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
		Layer:    util.MobLayer,
		Offset:   offset.Add(m.bob()),
		Drawable: context.GetDrawable(m.icon)}
}

func (m *Mob) BlocksMove() bool {
	return true
}

func (m *Mob) Health() int { return m.health }

func (m *Mob) MaxHealth() int { return m.maxHealth }

func (m *Mob) Shield() int { return m.shield }

func (m *Mob) Damage(amount int) {
	amountLeft := amount

	if m.shield > 0 {
		shieldDamage := num.MinI(m.shield, amountLeft)
		m.shield -= shieldDamage
		amountLeft -= shieldDamage

		// Freebie damage reduction from destroying shields. Only really
		// massive damage penetrates on the same turn.
		amountLeft = num.MaxI(0, amountLeft-3)
	}

	m.health = num.MaxI(0, m.health-amountLeft)
}

func (m *Mob) AddHealth(amount int) {
	m.health = num.ClampI(0, m.maxHealth, m.health+amount)
}

func (m *Mob) AddShield(amount int) {
	m.shield = num.MaxI(0, m.shield+amount)
}
