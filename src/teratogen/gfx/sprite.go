// sprite.go
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

package gfx

import (
	"image"
	"sort"
	"unsafe"
)

type Spritable interface {
	Sprite(context Context, offset image.Point) Sprite
}

type Sprite struct {
	Layer    int
	Drawable Drawable
	Offset   image.Point
}

func (s Sprite) Draw() {
	s.Drawable.Draw(s.Offset)
}

// For sorting sprites.

type SpriteBatch []Sprite

func (s SpriteBatch) Len() int { return len(s) }

func (s SpriteBatch) Less(i, j int) bool {
	if s[i].Layer != s[j].Layer {
		return s[i].Layer < s[j].Layer
	}
	// XXX: Do I really need to write out all these?
	if s[i].Offset.Y != s[j].Offset.Y {
		return s[i].Offset.Y < s[j].Offset.Y
	}
	if s[i].Offset.X != s[j].Offset.X {
		return s[i].Offset.X < s[j].Offset.X
	}
	return uintptr(unsafe.Pointer(&s[i].Drawable)) < uintptr(unsafe.Pointer(&s[j].Drawable))
}

func (s SpriteBatch) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Sort sorts a SpriteBatch according to Layer. It places identical sprites
func (s SpriteBatch) Sort() {
	sort.Sort(s)
}

// Draw draws a SpriteBatch that is assumed to be sorted. Duplicate sprites
// are skipped.
func (s SpriteBatch) Draw() {
	if len(s) == 0 {
		return
	}
	s[0].Draw()
	prev := s[0]
	for _, spr := range s[1:] {
		if spr == prev {
			continue
		}
		prev = spr
		spr.Draw()
	}
}
