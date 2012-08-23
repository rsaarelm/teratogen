// world.go
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

package world

import (
	"image"
	"teratogen/entity"
	"teratogen/fov"
	"teratogen/gfx"
	"teratogen/manifold"
	"teratogen/mapgen"
	"teratogen/spatial"
)

type World struct {
	Manifold *manifold.Manifold
	terrain  map[manifold.Location]Terrain
	Spatial  *spatial.Spatial

	Player interface {
		gfx.Spritable
		entity.Pos
	}
}

type WorldFormer struct {
	world *World
	chart manifold.Chart
}

func (w WorldFormer) At(p image.Point) mapgen.Terrain {
	loc := w.chart.At(p)
	if t, ok := w.world.terrain[loc]; ok {
		switch t {
		case WallTerrain:
			return mapgen.Solid
		case FloorTerrain:
			return mapgen.Open
		case DoorTerrain:
			return mapgen.Doorway
		}
	}

	return mapgen.Solid
}

func (w WorldFormer) Set(p image.Point, t mapgen.Terrain) {
	loc := w.chart.At(p)
	switch t {
	case mapgen.Solid:
		w.world.terrain[loc] = WallTerrain
	case mapgen.Open:
		w.world.terrain[loc] = FloorTerrain
	case mapgen.Doorway:
		w.world.terrain[loc] = DoorTerrain
	}
}

func New() (world *World) {
	world = new(World)
	world.Manifold = manifold.New()
	world.terrain = make(map[manifold.Location]Terrain)
	world.Spatial = spatial.New()
	return
}

func (w *World) TestMap(origin manifold.Location) {
	bounds := image.Rect(-16, -16, 16, 16)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			w.terrain[origin.Add(image.Pt(x, y))] = WallTerrain
		}
	}
	mapgen.BspRooms(WorldFormer{w, simpleChart(origin)}, bounds.Inset(1))
}

func (w *World) Terrain(loc manifold.Location) TerrainData {
	if t, ok := w.terrain[loc]; ok {
		return terrainTable[t]
	}
	return terrainTable[VoidTerrain]
}

func (w *World) Contains(loc manifold.Location) bool {
	_, ok := w.terrain[loc]
	return ok
}

// simpleChart is a chart that pays no attention to portals in the manifold.
type simpleChart manifold.Location

func (s simpleChart) At(pt image.Point) manifold.Location {
	return manifold.Location(s).Add(pt)
}

type FovChart struct {
	RelativePos image.Point
	world       *World
	chart       map[image.Point]manifold.Location
}

func NewFov(world *World) (result *FovChart) {
	result = new(FovChart)
	result.world = world
	result.chart = make(map[image.Point]manifold.Location)
	return
}

func (fc *FovChart) Move(vec image.Point) {
	fc.RelativePos = fc.RelativePos.Add(vec)
}

func (fc *FovChart) At(pt image.Point) manifold.Location {
	if loc, ok := fc.chart[pt.Add(fc.RelativePos)]; ok {
		return loc
	}
	return manifold.Location{}
}

func (fc *FovChart) Mark(pt image.Point, loc manifold.Location) {
	fc.chart[pt.Add(fc.RelativePos)] = loc
}

func (fc *FovChart) DoFov(origin manifold.Location, radius int) {
	f := fov.New(
		func(loc manifold.Location) bool { return fc.world.Terrain(loc).BlocksSight() },
		func(pt image.Point, loc manifold.Location) { fc.Mark(pt, loc) },
		fc.world.Manifold)
	f.Run(origin, radius)
}