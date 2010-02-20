package main

import (
	"hyades/alg"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"hyades/gfx"
	"io"
	"rand"
)

const mapWidth = 40
const mapHeight = 20

const numTerrainCells = mapWidth * mapHeight

// Behavioral terrain types.
type TerrainType byte

const (
	// Used for terrain generation algorithms, set map to indeterminate
	// initially.
	TerrainIndeterminate TerrainType = iota
	TerrainWallFront
	TerrainWall
	TerrainFloor
	TerrainDoor
	TerrainStairDown
	TerrainDirtFront
	TerrainDirt
)

var tileset1 = []string{
	TerrainIndeterminate: "tiles:255",
	TerrainWall: "tiles:2",
	TerrainWallFront: "tiles:1",
	TerrainFloor: "tiles:0",
	TerrainDoor: "tiles:3",
	TerrainStairDown: "tiles:4",
	TerrainDirt: "tiles:6",
	TerrainDirtFront: "tiles:5",
}

const AreaComponent = entity.ComponentFamily("area")

type Area struct {
	terrain []TerrainType
}

func NewArea() (result *Area) {
	result = new(Area)
	result.terrain = make([]TerrainType, numTerrainCells)
	return
}

func (self *Area) Serialize(out io.Writer) { entity.GobSerialize(out, self) }

func (self *Area) Deserialize(in io.Reader) { entity.GobDeserialize(in, self) }

func IsObstacleTerrain(terrain TerrainType) bool {
	switch terrain {
	case TerrainWall, TerrainDirt:
		return true
	}
	return false
}

func (self *Area) InArea(pos geom.Pt2I) bool {
	return pos.X >= 0 && pos.Y >= 0 && pos.X < mapWidth && pos.Y < mapHeight
}

func (self *Area) BlocksSight(pos geom.Pt2I) bool {
	if IsObstacleTerrain(self.GetTerrain(pos)) {
		return true
	}
	if self.GetTerrain(pos) == TerrainDoor {
		return true
	}

	return false
}

func (self *Area) GetTerrain(pos geom.Pt2I) TerrainType {
	if self.InArea(pos) {
		return self.terrain[pos.X+pos.Y*mapWidth]
	}
	return TerrainIndeterminate
}

func (self *Area) SetTerrain(pos geom.Pt2I, t TerrainType) {
	if self.InArea(pos) {
		self.terrain[pos.X+pos.Y*mapWidth] = t
	}
}

func (self *Area) MakeBSPMap() {
	area := MakeBspMap(1, 1, mapWidth-2, mapHeight-2)
	graph := alg.NewSparseMatrixGraph()
	area.FindConnectingWalls(graph)
	doors := DoorLocations(graph)

	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		x, y := pt.X, pt.Y
		if area.RoomAtPoint(x, y) != nil {
			self.SetTerrain(geom.Pt2I{x, y}, TerrainFloor)
		} else {
			self.SetTerrain(geom.Pt2I{x, y}, TerrainWall)
		}
	}

	for pt := range doors.Iter() {
		pt := pt.(geom.Pt2I)
		self.SetTerrain(pt, TerrainDoor)
	}
}

func (self *Area) MakeCaveMap() {
	area := MakeCaveMap(mapWidth, mapHeight, 0.50)
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		switch area[pt.X][pt.Y] {
		case CaveFloor:
			self.SetTerrain(pt, TerrainFloor)
		case CaveWall:
			self.SetTerrain(pt, TerrainDirt)
		case CaveUnknown:
			self.SetTerrain(pt, TerrainDirt)
		default:
			dbg.Die("Bad data %v in generated cave map.", area[pt.X][pt.Y])
		}
	}
}

func drawTerrain(g gfx.Graphics) {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		if GetLos().Get(pt) == LosUnknown {
			continue
		}
		idx := GetArea().GetTerrain(pt)
		front := GetArea().GetTerrain(pt.Plus(geom.Vec2I{0, 1}))
		// XXX: Hack to get the front tile visuals
		if idx == TerrainWall && front != TerrainWall && front != TerrainDoor {
			idx = TerrainWallFront
		}
		if idx == TerrainDirt && front != TerrainDirt && front != TerrainDoor {
			idx = TerrainDirtFront
		}
		Draw(g, tileset1[idx], pt.X, pt.Y)
	}
}

func IsOpen(pos geom.Pt2I) bool {
	if IsObstacleTerrain(GetArea().GetTerrain(pos)) {
		return false
	}
	for o := range EntitiesAt(pos).Iter() {
		id := o.(entity.Id)
		if ent := GetBlob(id); ent != nil {
			if ent.Has(FlagObstacle) {
				return false
			}
		}
	}

	return true
}

func GetSpawnPos() (pos geom.Pt2I) {
	pos, ok := GetMatchingPos(
		func(pos geom.Pt2I) bool { return isSpawnPos(pos) })
	// XXX: Maybe this shouldn't be an assert, since a situation where no
	// spawn pos can be found can occur during play.
	dbg.Assert(ok, "Couldn't find open spawn position.")
	return
}

func isSpawnPos(pos geom.Pt2I) bool {
	if !IsOpen(pos) {
		return false
	}
	if GetArea().GetTerrain(pos) == TerrainDoor {
		return false
	}
	if GetArea().GetTerrain(pos) == TerrainStairDown {
		return false
	}
	return true
}

func GetMatchingPos(f func(geom.Pt2I) bool) (pos geom.Pt2I, found bool) {
	const tries = 1024

	for i := 0; i < tries; i++ {
		x, y := rand.Intn(mapWidth), rand.Intn(mapHeight)
		pos = geom.Pt2I{x, y}
		if f(pos) {
			return pos, true
		}
	}

	// RNG has failed us, let's do an exhaustive search...
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		if f(pt) {
			return pt, true
		}
	}

	// There really doesn't seem to be any open positions.
	return geom.Pt2I{0, 0}, false
}
