package main

import (
	"container/vector"
	"fmt"
	"hyades/alg"
	"hyades/dbg"
	"hyades/geom"
	"hyades/num"
	"image"
	"rand"
)

const mapWidth = 40
const mapHeight = 20

const numTerrainCells = mapWidth * mapHeight

type Icon struct {
	IconId	string
	Color	image.Color
}

const xDrawOffset = 0
const yDrawOffset = 0

const TileW = 16
const TileH = 16

var world *World

func (self *Icon) Draw(x, y int) {
	DrawSprite(self.IconId, TileW * x + xDrawOffset, TileH * y + yDrawOffset)
}


// Behavioral terrain types.
type TerrainType byte

const (
	// Used for terrain generation algorithms, set map to indeterminate
	// initially.
	TerrainIndeterminate	= iota
	TerrainWall
	TerrainFloor
	TerrainDoor
	TerrainStairDown
)

type EntityType int

const (
	EntityUnknown	= iota
	EntityPlayer
	EntityZombie
	EntityBigboss
	EntityMinorHealthGlobe
)

// Put item classes before creature classes, so we can use this to control
// draw order as well.
type EntityClass int

const (
	EmptyEntityClass	= iota

	// Item classes
	GlobeEntityClass;	// Globe items are used when stepped on.

	// Creature classes
	CreatureEntityClassStartMarker

	PlayerEntityClass
	EnemyEntityClass
)

type LosState byte

const (
	LosUnknown	= iota
	LosMapped
	LosSeen
)

var tileset1 = []Icon{
TerrainIndeterminate: Icon{"tiles:115", image.RGBAColor{0xff, 0, 0xff, 0xff}},
	TerrainWall: Icon{"tiles:32", image.RGBAColor{0x55, 0x55, 0x55, 0xff}},
	TerrainFloor: Icon{"tiles:5", image.RGBAColor{0xaa, 0xaa, 0xaa, 0xff}},
	TerrainDoor: Icon{"tiles:210", image.RGBAColor{0x00, 0xcc, 0xcc, 0xff}},
	TerrainStairDown: Icon{"tiles:8", image.RGBAColor{0xff, 0xff, 0xff, 0xff}},
}

func IsObstacleTerrain(terrain TerrainType) bool {
	switch terrain {
	case TerrainWall:
		return true
	}
	return false
}

// Skinning data for a terrain tile set, describes the outward appearance of a
// type of terrain.
type TerrainTile struct {
	Icon
	Name	string
}


type Drawable interface {
	Draw(x, y int)
}


type Guid string


type Entity interface {
	Drawable
	// TODO: Entity-common stuff.
	IsObstacle() bool
	GetPos() geom.Pt2I
	GetGuid() Guid
	MoveAbs(pos geom.Pt2I)
	Move(vec geom.Vec2I)
	GetName() string
	GetClass() EntityClass
}


type World struct {
	playerId	Guid
	entities	map[Guid]Entity
	terrain		[]TerrainType
	los		[]LosState
	guidCounter	uint64
	currentLevel	int
}

func NewWorld() (result *World) {
	result = new(World)
	world = result
	result.entities = make(map[Guid]Entity)
	result.initTerrain()

	result.playerId = Guid("player")
	player := result.Spawn(EntityPlayer)
	result.playerId = player.GetGuid()

	result.InitLevel(1)

	return
}

func GetWorld() *World {
	dbg.AssertNotNil(world, "World not initialized.")
	return world
}

func (self *World) Draw() {
	self.drawTerrain()
	self.drawEntities()
}

func (self *World) GetPlayer() *Creature	{ return self.entities[self.playerId].(*Creature) }

func (self *World) GetEntity(guid Guid) (ent Entity, ok bool) {
	ent, ok = self.entities[guid]
	return
}

func (self *World) DestroyEntity(ent Entity) {
	if ent == Entity(self.GetPlayer()) {
		// TODO: End game when player dies.
		//		Msg("A mysterious anthropic effect prevents your discorporation.\n")
		//		ent.(*Creature).Wounds = 0
		GameOver("was wiped out of existence.")
		return
	}
	self.entities[ent.GetGuid()] = ent, false
}

func (self *World) Spawn(entityType EntityType) (result Entity) {
	guid := self.getGuid("")
	switch entityType {
	case EntityPlayer:
		result = &Creature{Icon: Icon{"guys:8", image.RGBAColor{0xdd, 0xff, 0xff, 0xff}},
			guid: guid,
			Name: "protagonist",
			pos: geom.Pt2I{-1, -1},
			class: PlayerEntityClass,
			// XXX: Give player superstrength until we get some weapons in play.
			Strength: Superb,
			Toughness: Good,
			MeleeSkill: Good,
		}
	case EntityZombie:
		result = &Creature{Icon: Icon{"guys:195", image.RGBAColor{0x80, 0xa0, 0x80, 0xff}},
			guid: guid,
			Name: "zombie",
			pos: geom.Pt2I{-1, -1},
			class: EnemyEntityClass,
			Strength: Fair,
			Toughness: Poor,
			MeleeSkill: Fair,
		}
	case EntityBigboss:
		result = &Creature{Icon: Icon{"guys:185", image.RGBAColor{0xa0, 0x00, 0xa0, 0xff}},
			guid: guid,
			Name: "elder spawn",
			pos: geom.Pt2I{-1, -1},
			class: EnemyEntityClass,
			Strength: Legendary,
			Toughness: Legendary,
			MeleeSkill: Superb,
			Scale: 15,
		}
	case EntityMinorHealthGlobe:
		result = &Item{Icon: Icon{"items:57", image.RGBAColor{0xff, 0x44, 0x44, 0xff}},
			guid: guid,
			Name: "health globe",
			pos: geom.Pt2I{-1, -1},
			class: GlobeEntityClass,
		}
	default:
		dbg.Die("Unknown entity type %v.", entityType)
	}
	self.entities[guid] = result
	return
}

func (self *World) SpawnAt(entityType EntityType, pos geom.Pt2I) (result Entity) {
	result = self.Spawn(entityType)
	result.MoveAbs(pos)
	return
}

func (self *World) SpawnRandomPos(entityType EntityType) (result Entity) {
	return self.SpawnAt(entityType, self.GetSpawnPos())
}

func (self *World) InitLevel(depth int) {
	// Keep the player around even though the other entities get munged.
	// TODO: When we start having inventories, keep the player's items too.
	player := self.GetPlayer()

	self.currentLevel = 1

	self.initTerrain()
	self.entities = make(map[Guid]Entity)
	self.entities[self.playerId] = player

	if num.WithProb(0.5) {
		self.makeCaveMap()
	} else {
		self.makeBSPMap()
	}

	self.SetTerrain(self.GetSpawnPos(), TerrainStairDown)

	player.MoveAbs(self.GetSpawnPos())
	self.DoLos(player.GetPos())
	for i := 0; i < 10+depth*4; i++ {
		self.SpawnRandomPos(EntityZombie)
	}

	for i := 0; i < 10; i++ {
		self.SpawnRandomPos(EntityMinorHealthGlobe)
	}
	//	self.SpawnRandomPos(EntityBigboss)
}

func (self *World) CurrentLevelNum() int	{ return self.currentLevel }

func (self *World) initTerrain() {
	self.terrain = make([]TerrainType, numTerrainCells)
	self.los = make([]LosState, numTerrainCells)
}

func (self *World) ClearLosSight() {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth*pt.Y
		if self.los[idx] == LosSeen {
			self.los[idx] = LosMapped
		}
	}
}

func (self *World) ClearLosMapped() {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth*pt.Y
		self.los[idx] = LosUnknown
	}
}

func (self *World) MarkSeen(pos geom.Pt2I) {
	if inTerrain(pos) {
		self.los[pos.X+pos.Y*mapWidth] = LosSeen
	}
}

func (self *World) GetLos(pos geom.Pt2I) LosState {
	if inTerrain(pos) {
		return self.los[pos.X+pos.Y*mapWidth]
	}
	return LosUnknown
}

func (self *World) DoLos(center geom.Pt2I) {
	const losRadius = 12

	blocks := func(vec geom.Vec2I) bool { return self.BlocksSight(center.Plus(vec)) }

	outOfRadius := func(vec geom.Vec2I) bool { return int(vec.Abs()) > losRadius }

	for pt := range geom.LineOfSight(blocks, outOfRadius) {
		self.MarkSeen(center.Plus(pt))
	}
}

func (self *World) BlocksSight(pos geom.Pt2I) bool {
	if IsObstacleTerrain(self.GetTerrain(pos)) {
		return true
	}
	if self.GetTerrain(pos) == TerrainDoor {
		return true
	}

	return false
}

func (self *World) makeBSPMap() {
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

func (self *World) makeCaveMap() {
	area := MakeCaveMap(mapWidth, mapHeight, 0.50)
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		switch area[pt.X][pt.Y] {
		case CaveFloor:
			self.SetTerrain(pt, TerrainFloor)
		case CaveWall:
			self.SetTerrain(pt, TerrainWall)
		case CaveUnknown:
			self.SetTerrain(pt, TerrainWall)
		default:
			dbg.Die("Bad data %v in generated cave map.", area[pt.X][pt.Y])
		}
	}
}

func inTerrain(pos geom.Pt2I) bool {
	return pos.X >= 0 && pos.Y >= 0 && pos.X < mapWidth && pos.Y < mapHeight
}

func (self *World) GetTerrain(pos geom.Pt2I) TerrainType {
	if inTerrain(pos) {
		return self.terrain[pos.X+pos.Y*mapWidth]
	}
	return TerrainIndeterminate
}

func (self *World) SetTerrain(pos geom.Pt2I, t TerrainType) {
	if inTerrain(pos) {
		self.terrain[pos.X+pos.Y*mapWidth] = t
	}
}

func (self *World) EntitiesAt(pos geom.Pt2I) <-chan Entity {
	c := make(chan Entity)
	go func() {
		for _, ent := range self.entities {
			if ent.GetPos().Equals(pos) {
				c <- ent
			}
		}
		close(c)
	}()
	return c
}

func (self *World) IterEntities() <-chan Entity {
	c := make(chan Entity)
	go func() {
		for _, ent := range self.entities {
			c <- ent
		}
		close(c)
	}()
	return c
}

func (self *World) IterCreatures() <-chan *Creature {
	c := make(chan *Creature)
	go func() {
		for _, ent := range self.entities {
			switch crit := ent.(type) {
			case *Creature:
				c <- crit
			}
		}
		close(c)
	}()
	return c
}


func (self *World) IsOpen(pos geom.Pt2I) bool {
	if IsObstacleTerrain(self.GetTerrain(pos)) {
		return false
	}
	for e := range self.EntitiesAt(pos) {
		if e.IsObstacle() {
			return false
		}
	}

	return true
}

func (self *World) GetSpawnPos() (pos geom.Pt2I) {
	pos, ok := self.GetMatchingPos(
		func(pos geom.Pt2I) bool { return self.isSpawnPos(pos) })
	// XXX: Maybe this shouldn't be an assert, since a situation where no
	// spawn pos can be found can occur during play.
	dbg.Assert(ok, "Couldn't find open spawn position.")
	return
}

func (self *World) isSpawnPos(pos geom.Pt2I) bool {
	if !self.IsOpen(pos) {
		return false
	}
	if self.GetTerrain(pos) == TerrainDoor {
		return false
	}
	if self.GetTerrain(pos) == TerrainStairDown {
		return false
	}
	return true
}

func (self *World) GetMatchingPos(f func(geom.Pt2I) bool) (pos geom.Pt2I, found bool) {
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


func (self *World) drawTerrain() {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		if self.GetLos(pt) == LosUnknown {
			continue
		}
		tileset1[self.GetTerrain(pt)].Draw(pt.X, pt.Y)
	}
}

func (self *World) drawEntities() {
	// Make a vector of the entities sorted in draw order.
	seq := new(vector.Vector)
	for e := range self.IterEntities() {
		seq.Push(e)
	}
	alg.PredicateSort(entityEarlierInDrawOrder, seq)

	for sorted := range seq.Iter() {
		e := sorted.(Entity)
		pos := e.GetPos()
		seen := self.GetLos(pos) == LosSeen
		mapped := seen || self.GetLos(pos) == LosMapped
		// TODO: Draw static (item) entities from map memory.
		if mapped {
			if seen || !IsMobile(e) {
				e.Draw(pos.X, pos.Y)
			}
		}
	}
}

func entityEarlierInDrawOrder(i, j interface{}) bool {
	return i.(Entity).GetClass() < j.(Entity).GetClass()
}

func (self *World) getGuid(name string) (result Guid) {
	result = Guid(fmt.Sprintf("%v#%v", name, self.guidCounter))
	self.guidCounter++
	return
}
