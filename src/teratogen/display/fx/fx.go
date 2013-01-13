// fx.go
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

// Package fx provides a high-level API for invoking game events that are
// converted into user interface visual effects.
package fx

import (
	"image"
	"teratogen/display/anim"
	"teratogen/space"
	"teratogen/world"
)

type BeamKind uint8

const (
	GunBeam BeamKind = iota
	ElectroBeam
	ContrailBeam
	FlameBeam
)

type BlastKind uint8

const (
	SparkBlast BlastKind = iota
	SmokeBlast
	ExplodeBlast
	WarpBlast
)

type Fx struct {
	anim  *anim.Anim
	world *world.World
}

func New(a *anim.Anim, w *world.World) *Fx {
	return &Fx{anim: a, world: w}
}

// Msgf prints a formatted message to the 
func (f *Fx) Msgf(format string, a ...interface{}) {
	// TODO: Get messaging system attached, deploy there.
}

// SpaceMsgf generates a message popup over a location in the game world.
func (f *Fx) SpaceMsgf(loc space.Location, format string, a ...interface{}) {
	// TODO
}

// Beam generates a projectile beam effect in the game world.
func (f *Fx) Beam(origin space.Location, dir image.Point, length int, kind BeamKind) {

}

// Blast generates an explosion effect in the game world.
func (f *Fx) Blast(loc space.Location, kind BlastKind) {
	// TODO
}
