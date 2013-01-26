// terrain.go
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
	"teratogen/display/util"
	"teratogen/gfx"
)

type Terrain uint8

type TerrainData struct {
	Icon []gfx.ImageSpec
	Kind TerrainKind
}

type TerrainKind uint8

const (
	SolidKind TerrainKind = iota
	WallKind
	OpenKind
	DoorKind
	GrillKind
	StairKind
)

func (t TerrainData) ShapesWalls() bool {
	return t.Kind == WallKind || t.Kind == DoorKind
}

func (t TerrainData) BlocksSight() bool {
	switch t.Kind {
	case SolidKind, WallKind, DoorKind:
		return true
	}
	return false
}

func (t TerrainData) BlocksMove() bool {
	switch t.Kind {
	case SolidKind, WallKind, GrillKind:
		return true
	}
	return false
}

const (
	VoidTerrain Terrain = iota
	FloorTerrain
	WallTerrain
	DoorTerrain
	StairTerrain
)

func tiles(idxs ...int) (result []gfx.ImageSpec) {
	for _, n := range idxs {
		result = append(result, util.SmallIcon(util.Tiles, n))
	}
	return
}

var terrainTable = []TerrainData{
	{tiles(3), SolidKind}, // void terrain, should have some "you shouldn't be seeing this" icon
	{tiles(0), OpenKind},
	{tiles(16, 17, 18, 19), WallKind},
	{tiles(3), DoorKind},
	{tiles(4), StairKind},
}
