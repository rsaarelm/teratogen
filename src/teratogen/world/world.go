/* world.go

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

package world

import (
	"image"
	"teratogen/cache"
	"teratogen/manifold"
	"teratogen/mapgen"
)

type Terrain uint8

type TerrainData struct {
	Icon []cache.ImageSpec
	Kind TerrainKind
}

type TerrainKind uint8

const (
	SolidKind TerrainKind = iota
	WallKind
	OpenKind
	DoorKind
	GrillKind
)

const (
	VoidTerrain Terrain = iota
	FloorTerrain
	WallTerrain
	DoorTerrain
)

func tile(idx int) cache.ImageSpec {
	x, y := idx%16, idx/16
	const dim = 8
	return cache.ImageSpec{"assets/tiles.png", image.Rect(x*dim, y*dim, x*dim+dim, y*dim+dim)}
}

func tiles(idxs ...int) (result []cache.ImageSpec) {
	for _, n := range idxs {
		result = append(result, tile(n))
	}
	return
}

var terrainTable = []TerrainData{
	{tiles(), SolidKind},
	{tiles(0), OpenKind},
	{tiles(16, 17, 18, 19), WallKind},
	{tiles(3), DoorKind},
}

type World struct {
	Manifold *manifold.Manifold
	Terrain  map[manifold.Location]Terrain
}

type WorldFormer struct {
	world *World
	chart manifold.Chart
}

func (w WorldFormer) At(p image.Point) mapgen.Terrain {
	loc := w.chart.At(p)
	if t, ok := w.world.Terrain[loc]; ok {
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
		w.world.Terrain[loc] = WallTerrain
	case mapgen.Open:
		w.world.Terrain[loc] = FloorTerrain
	case mapgen.Doorway:
		w.world.Terrain[loc] = DoorTerrain
	}
}

func New() (world *World) {
	world = new(World)
	world.Manifold = manifold.New()
	world.Terrain = make(map[manifold.Location]Terrain)
	return
}

func (w *World) TestMap(origin manifold.Location) {
	mapgen.BspRooms(WorldFormer{w, simpleChart(origin)}, image.Rect(-16, -16, 16, 16))
}

// simpleChart is a chart that pays no attention to portals in the manifold.
type simpleChart manifold.Location

func (s simpleChart) At(pt image.Point) manifold.Location {
	return manifold.Location(s).Add(pt)
}
