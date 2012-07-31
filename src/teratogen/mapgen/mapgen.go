/* mapgen.go

   Copyright (C) 2012 Risto Saarelma

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package mapgen

import (
	"image"
	"math"
	"math/rand"
)

type Terrain int

const (
	Solid Terrain = iota
	Open
	Doorway
)

type Former interface {
	At(p image.Point) Terrain
	Set(p image.Point, t Terrain)
}

// BspRooms builds an indoor rooms style map using a binary split tree
// algorithm.
func BspRooms(t Former, bounds image.Rectangle) {
	if bounds.Dx() < 1 || bounds.Dy() < 1 {
		return
	}

	const minArea = 8
	const maxArea = 128

	area := bounds.Dx() * bounds.Dy()

	if rand.Float64()*float64(maxArea-minArea)+float64(minArea) < float64(area) {
		splitRoom(t, bounds)
		return
	}

	digRoom(t, bounds)
}

func splitRoom(t Former, bounds image.Rectangle) {
	wall := makeSplitWall(bounds)
	left, right := wall.Halves(bounds)
	BspRooms(t, left)
	BspRooms(t, right)
	// TODO: Make a doorway along the wall.
}

type wall struct {
	Begin, End image.Point
}

func (w wall) IsVertical() bool {
	return w.Begin.X == w.End.X
}

func (w wall) Length() int {
	if w.IsVertical() {
		return w.End.Y - w.Begin.Y
	}
	return w.End.X - w.Begin.X
}

func (w wall) Halves(parent image.Rectangle) (left, right image.Rectangle) {
	if w.IsVertical() {
		left = image.Rect(parent.Min.X, parent.Min.Y, w.Begin.X, parent.Max.Y)
		right = image.Rect(w.Begin.X+1, parent.Min.Y, parent.Max.X, parent.Max.Y)
	} else {
		left = image.Rect(parent.Min.X, parent.Min.Y, parent.Max.X, w.Begin.Y)
		right = image.Rect(parent.Min.X, w.Begin.Y+1, parent.Max.X, parent.Max.Y)
	}
	return
}

// makeSplitWall picks a wall to split a room with, and returns a
// specification of the wall.
func makeSplitWall(bounds image.Rectangle) wall {
	vertWeight := int(math.Max(0, float64(bounds.Dx()-3)))
	horzWeight := int(math.Max(0, float64(bounds.Dy()-3)))
	isVertical := rand.Intn(vertWeight+horzWeight) < vertWeight

	if isVertical {
		offset := rand.Intn(bounds.Dx()-2) + 1
		return wall{
			image.Pt(bounds.Min.X+offset, bounds.Min.Y),
			image.Pt(bounds.Min.X+offset, bounds.Max.Y)}
	}

	offset := rand.Intn(bounds.Dy()-2) + 1
	return wall{
		image.Pt(bounds.Min.X, bounds.Min.Y+offset),
		image.Pt(bounds.Max.X, bounds.Min.Y+offset)}
}

func digRoom(t Former, bounds image.Rectangle) {
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			t.Set(image.Pt(x, y), Open)
		}
	}
}
