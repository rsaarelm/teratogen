package main

import (
	"exp/iterable"
	"hyades/alg"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"hyades/gfx"
	"hyades/mem"
	"io"
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

type AreaHandler struct {
	components map[entity.Id]*Area
}

func (self *AreaHandler) Init() {
	self.components = make(map[entity.Id]*Area)
	return
}

func (self *AreaHandler) Add(guid entity.Id, component interface{}) {
	self.components[guid] = component.(*Area)
}

func (self *AreaHandler) Remove(guid entity.Id) {
	self.components[guid] = nil, false
}

func (self *AreaHandler) Get(guid entity.Id) interface{} {
	if result, ok := self.components[guid]; ok {
		return result
	}
	return nil
}

func (self *AreaHandler) Serialize(out io.Writer) {
	entity.SerializeHandlerComponents(out, self)
}

func (self *AreaHandler) Deserialize(in io.Reader) {
	self.Init()
	entity.DeserializeHandlerComponents(in, self, mem.BlankCopier(new(Area)))
}

func (self *AreaHandler) EntityComponents() iterable.Iterable {
	return alg.IterFunc(func(c chan<- interface{}) {
		for id, comp := range self.components {
			c <- &entity.IdComponent{id, comp}
		}
		close(c)
	})
}

type Area struct {
	terrain []TerrainType
}

func NewArea() (result *Area) {
	result = new(Area)
	result.terrain = make([]TerrainType, numTerrainCells)
	return
}

func (self *Area) Serialize(out io.Writer) {
	mem.WriteNTimes(out, len(self.terrain), func(i int, out io.Writer) { mem.WriteFixed(out, byte(self.terrain[i])) })
}

func (self *Area) Deserialize(in io.Reader) {
	mem.ReadNTimes(in,
		func(count int) { self.terrain = make([]TerrainType, count) },
		func(i int, in io.Reader) { self.terrain[i] = TerrainType(mem.ReadByte(in)) })
}

// TODO: Move methods from World to Area.

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

func (self *World) drawTerrain(g gfx.Graphics) {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		if self.GetLos(pt) == LosUnknown {
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
