// bsp.go
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

package mapgen

import (
	"image"
	"math"
	"math/rand"
	"teratogen/world"
)

func (m *Mapgen) bspRooms(bounds image.Rectangle) {
	if bounds.Dx() < 1 || bounds.Dy() < 1 {
		return
	}

	const minArea = 8
	const maxArea = 96

	area := bounds.Dx() * bounds.Dy()

	if rand.Float64()*float64(maxArea-minArea)+float64(minArea) < float64(area) {
		m.splitRoom(bounds)
		return
	}

	m.digRoom(bounds)
}

func (m *Mapgen) extraDoors(bounds image.Rectangle) {
	const extraDoorChance = 1.0 / 128

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pt := image.Pt(x, y)
			if m.isDoorSite(pt) && rand.Float64() < extraDoorChance {
				m.setTerrain(pt, world.DoorTerrain)
			}
		}
	}

}

func (m *Mapgen) isDoorSite(pt image.Point) bool {
	if !m.terrain(pt).BlocksMove() {
		return false
	}

	up := m.terrain(pt.Add(image.Pt(0, -1))).BlocksMove()
	down := m.terrain(pt.Add(image.Pt(0, 1))).BlocksMove()
	left := m.terrain(pt.Add(image.Pt(-1, 0))).BlocksMove()
	right := m.terrain(pt.Add(image.Pt(1, 0))).BlocksMove()

	if up != down || left != right {
		return false
	}

	return up != left
}

func (m *Mapgen) digRoom(bounds image.Rectangle) {
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pos := image.Pt(x, y)
			m.setTerrain(pos, world.FloorTerrain)
			m.setOpen(m.chart.At(pos), true)
		}
	}
}

func (m *Mapgen) splitRoom(bounds image.Rectangle) {
	wall := makeSplitWall(bounds)
	left, right := wall.Halves(bounds)
	m.bspRooms(left)
	m.bspRooms(right)

	doorSites := m.doorSites(wall)
	m.setTerrain(doorSites[rand.Intn(len(doorSites))], world.DoorTerrain)
}

// DoorSites returns points along the wall which are suitable for placing a
// doorway. Such points are ones which have open space on both of their sides.
func (m *Mapgen) doorSites(w wall) (result []image.Point) {
	result = []image.Point{}
	for pt := w.Begin; pt != w.End; pt = pt.Add(w.Dir()) {
		left, right := w.Sides()
		if !m.terrain(pt.Add(left)).BlocksMove() && !m.terrain(pt.Add(right)).BlocksMove() {
			result = append(result, pt)
		}
	}
	return
}

// makeSplitWall picks a wall to split a room with, and returns a
// specification of the wall.
func makeSplitWall(bounds image.Rectangle) wall {
	vertWeight := int(math.Max(0, float64(bounds.Dx()-3)))
	horzWeight := int(math.Max(0, float64(bounds.Dy()-3)))

	isVertical := false
	if horzWeight > 0 && vertWeight > 0 {
		isVertical = rand.Intn(vertWeight+horzWeight) < vertWeight
	} else if vertWeight > 0 {
		isVertical = true
	}

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

func (w wall) Dir() image.Point {
	if w.IsVertical() {
		return image.Pt(0, 1)
	}
	return image.Pt(1, 0)
}

func (w wall) Sides() (left, right image.Point) {
	if w.IsVertical() {
		return image.Pt(-1, 0), image.Pt(1, 0)
	}
	return image.Pt(0, -1), image.Pt(0, 1)
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
