package teratogen

import (
	"exp/iterable"
	"hyades/alg"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"hyades/num"
	"rand"
)

const mapWidth = 32
const mapHeight = 32

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

func (self *Area) AtEdge(pos geom.Pt2I) bool {
	return geom.IsAtEdge(0, 0, mapWidth, mapHeight, pos)
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

	self.placeCorridorDoors(0.80)
}

func (self *Area) placeCorridorDoors(prob float64) {
	dbg.Assert(num.IsProb(prob), "Bad corridor door prob %v", prob)

	for pos := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		if self.GetTerrain(pos) == TerrainCorridor {
			opensToRoom := false
			for i := 0; i < 6; i++ {
				if self.GetTerrain(pos.Plus(geom.Dir6ToVec(i))) == TerrainFloor {
					opensToRoom = true
					break
				}
			}
			if opensToRoom && num.WithProb(prob) {
				self.SetTerrain(pos, TerrainDoor)
			}
		}
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
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		if self.AtEdge(pt) {
			self.SetTerrain(pt, TerrainRockWall)
		} else {
			self.SetTerrain(pt, TerrainFloor)
		}
	}

	area := (mapWidth - 2) * (mapHeight - 2)
	areaToLeave := int(area / 8)

	// TODO: Add some walls at the edges, make it a bit more decayed-lookieng.

	for {
		buildingArea, ok := self.placeRuinBuilding()
		if !ok {
			break
		}
		area -= buildingArea
		if area <= areaToLeave {
			break
		}
	}
}

func (self *Area) placeRuinBuilding() (area int, success bool) {
	for nTries := 0; nTries < 32; nTries++ {
		w, h := rand.Intn(8)+3, rand.Intn(8)+3
		x0, y0 := rand.Intn(mapWidth-2-w)+2, rand.Intn(mapHeight-2-h)
		isSolid := num.OneChanceIn(5)
		edgeClear := iterable.All(geom.EdgeIter(x0-1, y0-1, w+2, h+2),
			func(o interface{}) bool { return self.GetTerrain(o.(geom.Pt2I)) == TerrainFloor })
		// Check the inside edge too.
		edgeClear = edgeClear && iterable.All(geom.EdgeIter(x0+1, y0+1, w-2, h-2),
			func(o interface{}) bool { return self.GetTerrain(o.(geom.Pt2I)) == TerrainFloor })
		if edgeClear {
			success = true
			area = w * h
			for pt := range geom.PtIter(x0, y0, w, h) {
				if geom.IsAtEdge(x0, y0, w, h, pt) {
					self.SetTerrain(pt, TerrainRockWall)
				} else if isSolid {
					self.SetTerrain(pt, TerrainRockWall)
				}
			}

			if !isSolid {
				// Carve at least one doorway
				doorway := TerrainCorridor
				if num.OneChanceIn(3) {
					doorway = TerrainDoor
				}
				points := iterable.Data(iterable.Filter(
					geom.EdgeIter(x0, y0, w, h),
					func(o interface{}) bool { return !geom.IsCorner(x0, y0, w, h, o.(geom.Pt2I)) }))
				for i, j := 0, rand.Intn(3)+1; i < j; i++ {
					self.SetTerrain(num.RandomChoiceA(points).(geom.Pt2I), doorway)
				}
			}
			return
		}
	}
	return
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

func IsBlocked(pos geom.Pt2I) bool {
	return !IsOpen(pos)
}

func BlocksRanged(pos geom.Pt2I) bool {
	return IsObstacleTerrain(GetArea().GetTerrain(pos))
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
