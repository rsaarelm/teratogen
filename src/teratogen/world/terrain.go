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
	ObstacleKind
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
	case SolidKind, WallKind, GrillKind, ObstacleKind:
		return true
	}
	return false
}

func (t TerrainData) BlocksShot() bool {
	switch t.Kind {
	case SolidKind, WallKind, DoorKind, ObstacleKind:
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

	BarrelTerrain
	ShelfTerrain
	ChairTerrain
	CounterTerrain
	PlantTerrain
)

func GetTerrainData(t Terrain) TerrainData {
	return terrainTable[t]
}

var terrainTable = []TerrainData{
	{util.SmallIcons(util.Tiles, 3), SolidKind}, // void terrain, should have some "you shouldn't be seeing this" icon
	{util.IsoIcons(util.Tiles, 5), OpenKind},
	{util.IsoIcons(util.Tiles, 1, 2, 3, 4), WallKind},
	{util.IsoIcons(util.Tiles, 7, 8, 9, 7), DoorKind},
	{util.IsoIcons(util.Tiles, 6), OpenKind},

	{util.IsoIcons(util.Tiles, 10), ObstacleKind},
	{util.IsoIcons(util.Tiles, 11), GrillKind},
	{util.IsoIcons(util.Tiles, 12), OpenKind},
	{util.IsoIcons(util.Tiles, 13), GrillKind},
	{util.IsoIcons(util.Tiles, 14), OpenKind},
}
