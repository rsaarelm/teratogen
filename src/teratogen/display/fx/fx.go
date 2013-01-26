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
	"teratogen/cache"
	"teratogen/display/anim"
	"teratogen/display/util"
	"teratogen/gfx"
	"teratogen/sdl"
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
	SmallExplodeBlast
	LargeExplodeBlast
	WarpBlast
)

type Fx struct {
	cache *cache.Cache
	anim  *anim.Anim
	world *world.World
}

func New(c *cache.Cache, a *anim.Anim, w *world.World) *Fx {
	return &Fx{cache: c, anim: a, world: w}
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
	// Make a footprint for the beam shape.
	shape := []image.Point{image.Pt(0, 0)}
	for i := 0; i < length; i++ {
		shape = append(shape, shape[len(shape)-1].Add(dir))
	}
	footprint := space.FootprintFromPoints(f.world.Manifold, origin, shape)

	screenVec := util.ChartToScreen(shape[len(shape)-1])

	// TODO: Different beam types.

	f.anim.Add(
		anim.Func(func(t int64, offset image.Point) {
			gfx.Line(
				sdl.Frame(),
				offset.Add(util.HalfTile),
				offset.Add(util.HalfTile).Add(screenVec),
				gfx.LerpCol(gfx.Gold, gfx.Black, float64(t)/float64(.5e9)))
		}), footprint, .2e9)
}

// Blast generates an explosion effect in the game world.
func (f *Fx) Blast(loc space.Location, kind BlastKind) {
	switch kind {
	case SmallExplodeBlast:
		frames := anim.NewCycle(f.cache, .1e9, false, util.SmallIcons(util.Items, 32, 33, 34, 35))
		f.anim.Add(
			anim.Func(func(t int64, offset image.Point) {
				frames.Frame(t).Draw(offset)
			}), space.SimpleFootprint(loc), .4e9)
	case LargeExplodeBlast:
		frames := anim.NewCycle(f.cache, .15e9, false, util.LargeIcons(util.Items, 5, 6, 7))
		f.anim.Add(
			anim.Func(func(t int64, offset image.Point) {
				frames.Frame(t).Draw(offset)
			}), space.SimpleFootprint(loc), .4e9)
	default:
		println("Unknown blast kind ", kind)
		return
	}

}
