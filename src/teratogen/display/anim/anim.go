// anim.go
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

// Package anim contains display logic for transient effect animations in the
// game view.
package anim

import (
	"image"
	"teratogen/app"
	"teratogen/display/util"
	"teratogen/gfx"
	"teratogen/space"
	"time"
)

type Animation interface {
	DrawFrame(time int64, offset image.Point)
}

func Func(f func(int64, image.Point)) *funcAnim {
	result := funcAnim(f)
	return &result
}

type funcAnim func(int64, image.Point)

func (f *funcAnim) DrawFrame(time int64, offset image.Point) {
	(*f)(time, offset)
}

// AnimationFrame is a gfx.Drawable that encapsulates a specific point in an
// animation. It can be wrapped into sprites generated for a frame.
type animationFrame struct {
	time      int64
	animation Animation
}

func (af animationFrame) Draw(offset image.Point) {
	af.animation.DrawFrame(af.time, offset)
}

func now() int64 {
	return time.Now().UnixNano()
}

type animationStore struct {
	createTime  int64
	destroyTime int64
	anim        Animation
}

func (a animationStore) IsDead() bool {
	return now() >= a.destroyTime
}

func (a animationStore) CurrentFrame() animationFrame {
	return animationFrame{now() - a.createTime, a.anim}
}

type Anim struct {
	index *space.Index
}

func New() (result *Anim) {
	result = new(Anim)
	result.index = space.NewIndex()
	return
}

func (a *Anim) Add(animation Animation, foot space.Footprint, duration int64) {
	t := now()
	obj := animationStore{t, t + duration, animation}
	a.index.Place(obj, foot)
}

func (a *Anim) CollectSpritesAt(
	sprites gfx.SpriteBatch,
	loc space.Location,
	offset image.Point,
	layer int) gfx.SpriteBatch {
	for _, oe := range a.index.At(loc) {
		screenPos := util.ChartToScreen(oe.Offset.Mul(-1)).Add(offset)
		animStore := oe.Entity.(animationStore)

		// Delete ended animations as we encounter them.
		if animStore.IsDead() {
			a.index.Remove(oe.Entity)
			continue
		}

		// Create sprites from the current frames of live animations.
		sprites = append(
			sprites,
			gfx.Sprite{layer, screenPos, animStore.CurrentFrame()})
	}
	return sprites
}

type Cycle struct {
	TimePerFrame int64
	Frames       []gfx.Drawable
	Loops        bool
}

func (c Cycle) Frame(t int64) gfx.Drawable {
	if c.TimePerFrame <= 0 {
		panic("Invalid Cycle")
	}
	idx := int(t / c.TimePerFrame)
	if idx >= len(c.Frames) {
		if c.Loops {
			idx %= len(c.Frames)
		} else {
			idx = len(c.Frames) - 1
		}
	}
	return c.Frames[idx]
}

func NewCycle(timePerFrame int64, loops bool, frameSpecs []gfx.ImageSpec) Cycle {
	var frames []gfx.Drawable
	for _, spec := range frameSpecs {
		frames = append(frames, app.Cache().GetDrawable(spec))
	}

	return Cycle{timePerFrame, frames, loops}
}
