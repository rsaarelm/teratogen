package teratogen

import (
	"hyades/alg"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"rand"
)

const mapWidth = 40
const mapHeight = 20

func MapDims() (width, height int) { return mapWidth, mapHeight }

const numTerrainCells = mapWidth * mapHeight

// Behavioral terrain types.
type TerrainType byte

const (
	// Used for terrain generation algorithms, set map to indeterminate
	// initially.
	TerrainIndeterminate TerrainType = iota
	TerrainFloor
	TerrainStairDown
	TerrainCorridor
	// Tiles after this are visual walls.
	TerrainDoor
	// Tiles after this are actual walls.
	TerrainWall
	TerrainBrickWall
	TerrainDirtWall
	TerrainRockWall
	TerrainBioWall
)

const AreaComponent = entity.ComponentFamily("area")

type Area struct {
	terrain []TerrainType
}

func NewArea() (result *Area) {
	result = new(Area)
	result.terrain = make([]TerrainType, numTerrainCells)
	return
}

func IsObstacleTerrain(terrain TerrainType) bool {
	switch {
	case terrain >= TerrainWall:
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
		if area.RoomAtPoint(pt.X, pt.Y) != nil {
			self.SetTerrain(pt, TerrainFloor)
		} else {
			self.SetTerrain(pt, TerrainWall)
		}
	}

	for pt := range doors.Iter() {
		pt := pt.(geom.Pt2I)
		self.SetTerrain(pt, TerrainDoor)
	}
}

type corridorDiggable Area

type roomDiggable Area

func (self *corridorDiggable) CanDig(pos geom.Pt2I) bool {
	return pos.X > 1 && pos.Y > 1 && pos.X < mapWidth-1 && pos.Y < mapHeight-1
}

func (self *roomDiggable) CanDig(pos geom.Pt2I) bool {
	return pos.X > 1 && pos.Y > 1 && pos.X < mapWidth-1 && pos.Y < mapHeight-1
}

func (self *corridorDiggable) IsDug(pos geom.Pt2I) bool {
	return (*Area)(self).GetTerrain(pos) == TerrainCorridor || (*Area)(self).GetTerrain(pos) == TerrainFloor
}

func (self *roomDiggable) IsDug(pos geom.Pt2I) bool {
	return (*Area)(self).GetTerrain(pos) == TerrainCorridor || (*Area)(self).GetTerrain(pos) == TerrainFloor
}

func (self *corridorDiggable) Dig(pos geom.Pt2I) {
	(*Area)(self).SetTerrain(pos, TerrainCorridor)
}

func (self *roomDiggable) Dig(pos geom.Pt2I) { (*Area)(self).SetTerrain(pos, TerrainFloor) }

func (self *Area) MakeCellarMap() {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		self.SetTerrain(pt, TerrainBrickWall)
	}

	corrDug := 0
	const needCorrDug = 200
	const needTotalDug = 600

	for nTries := 256; nTries > 0 && corrDug < needCorrDug; nTries-- {
		pos, ok := GetSpawnPos()
		if !ok {
			pos = geom.Pt2I{mapWidth / 2, mapHeight / 2}
		}
		corrDug += DigTunnels(pos, (*corridorDiggable)(self), 0.1, 0.05, 0.01)
	}

	roomDug := 0

	for nTries := 256; nTries > 0 && corrDug+roomDug < needTotalDug; nTries-- {
		roomDug += DigRoom((*roomDiggable)(self), 0, 0, mapWidth, mapHeight, 12, 12)
	}
}

func (self *Area) MakeCaveMap() {
	area := MakeHexCaveMap(mapWidth, mapHeight, 0.50)
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		switch area[pt.X][pt.Y] {
		case CaveFloor:
			self.SetTerrain(pt, TerrainFloor)
		case CaveWall:
			self.SetTerrain(pt, TerrainDirtWall)
		case CaveUnknown:
			self.SetTerrain(pt, TerrainDirtWall)
		default:
			dbg.Die("Bad data %v in generated cave map.", area[pt.X][pt.Y])
		}
	}
}

func (self *Area) MakeRuinsMap() {
	// TODO: Open area and buildings gen.
	area := MakeBspMap(1, 1, mapWidth-2-mapHeight, mapHeight-2)
	graph := alg.NewSparseMatrixGraph()
	area.FindConnectingWalls(graph)
	doors := DoorLocations(graph)

	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		x, y := pt.X, pt.Y
		bX, bY := x+y-mapHeight, y
		if area.RoomAtPoint(bX, bY) != nil {
			self.SetTerrain(geom.Pt2I{x, y}, TerrainFloor)
		} else {
			self.SetTerrain(geom.Pt2I{x, y}, TerrainRockWall)
		}
	}

	for pt := range doors.Iter() {
		pt := pt.(geom.Pt2I)
		x, y := pt.X-pt.Y+mapHeight, pt.Y
		self.SetTerrain(geom.Pt2I{x, y}, TerrainDoor)
	}
}

func (self *Area) MakeVisceraMap() {
	area := MakeHexCaveMap(mapWidth, mapHeight, 0.50)
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		switch area[pt.X][pt.Y] {
		case CaveFloor:
			self.SetTerrain(pt, TerrainFloor)
		case CaveWall:
			self.SetTerrain(pt, TerrainBioWall)
		case CaveUnknown:
			self.SetTerrain(pt, TerrainBioWall)
		default:
			dbg.Die("Bad data %v in generated cave map.", area[pt.X][pt.Y])
		}
	}
}


// IsUnwalkable returns whether the terrain in pos can't be walked into.
func IsUnwalkable(pos geom.Pt2I) bool { return IsObstacleTerrain(GetArea().GetTerrain(pos)) }

func IsOpen(pos geom.Pt2I) bool {
	if IsObstacleTerrain(GetArea().GetTerrain(pos)) {
		return false
	}
	for o := range EntitiesAt(pos).Iter() {
		id := o.(entity.Id)
		if BlocksMovement(id) {
			return false
		}
	}

	return true
}

func GetSpawnPos() (pos geom.Pt2I, ok bool) {
	return GetMatchingPos(
		func(pos geom.Pt2I) bool { return isSpawnPos(pos) })
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
