package main

import "container/vector"
import "fmt"
import "rand"

import . "hyades/gamelib"

const mapWidth = 40
const mapHeight = 20

const numTerrainCells = mapWidth * mapHeight

type Icon struct {
	IconId	byte
	Color	RGB
}

const xDrawOffset = 0
const yDrawOffset = 0

var world *World

func (self *Icon) Draw(x, y int) {
	GetConsole().SetCF(
		x+xDrawOffset, y+yDrawOffset,
		int(self.IconId), self.Color)
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
	TerrainIndeterminate: Icon{'?', RGB{0xff, 0, 0xff}},
	TerrainWall: Icon{'#', RGB{0x55, 0x55, 0x55}},
	TerrainFloor: Icon{'.', RGB{0xaa, 0xaa, 0xaa}},
	TerrainDoor: Icon{'+', RGB{0x00, 0xcc, 0xcc}},
	TerrainStairDown: Icon{'>', RGB{0xff, 0xff, 0xff}},
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
	GetPos() Pt2I
	GetGuid() Guid
	MoveAbs(pos Pt2I)
	Move(vec Vec2I)
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
	AssertNotNil(world, "World not initialized.")
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
		result = &Creature{Icon: Icon{'@', RGB{0xdd, 0xff, 0xff}},
			guid: guid,
			Name: "protagonist",
			pos: Pt2I{-1, -1},
			class: PlayerEntityClass,
			// XXX: Give player superstrength until we get some weapons in play.
			Strength: Superb,
			Toughness: Good,
			MeleeSkill: Good,
		}
	case EntityZombie:
		result = &Creature{Icon: Icon{'z', RGB{0x80, 0xa0, 0x80}},
			guid: guid,
			Name: "zombie",
			pos: Pt2I{-1, -1},
			class: EnemyEntityClass,
			Strength: Fair,
			Toughness: Poor,
			MeleeSkill: Fair,
		}
	case EntityBigboss:
		result = &Creature{Icon: Icon{'Q', RGB{0xa0, 0x00, 0xa0}},
			guid: guid,
			Name: "elder spawn",
			pos: Pt2I{-1, -1},
			class: EnemyEntityClass,
			Strength: Legendary,
			Toughness: Legendary,
			MeleeSkill: Superb,
			Scale: 15,
		}
	case EntityMinorHealthGlobe:
		result = &Item{Icon: Icon{'%', RGB{0xff, 0x44, 0x44}},
			guid: guid,
			Name: "health globe",
			pos: Pt2I{-1, -1},
			class: GlobeEntityClass,
		}
	default:
		Die("Unknown entity type %v.", entityType)
	}
	self.entities[guid] = result
	return
}

func (self *World) SpawnAt(entityType EntityType, pos Pt2I) (result Entity) {
	result = self.Spawn(entityType)
	result.MoveAbs(pos)
	return
}

func (self *World) SpawnRandomPos(entityType EntityType) (result Entity) {
	return self.SpawnAt(entityType, self.GetSpawnPos())
}

func (self *World) InitLevel(num int) {
	// Keep the player around even though the other entities get munged.
	// TODO: When we start having inventories, keep the player's items too.
	player := self.GetPlayer()

	self.currentLevel = 1

	self.initTerrain()
	self.entities = make(map[Guid]Entity)
	self.entities[self.playerId] = player

	if WithProb(0.5) {
		self.makeCaveMap()
	} else {
		self.makeBSPMap()
	}

	self.SetTerrain(self.GetSpawnPos(), TerrainStairDown)

	player.MoveAbs(self.GetSpawnPos())
	self.DoLos(player.GetPos())
	for i := 0; i < 10+num*4; i++ {
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
	for pt := range PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth*pt.Y
		if self.los[idx] == LosSeen {
			self.los[idx] = LosMapped
		}
	}
}

func (self *World) ClearLosMapped() {
	for pt := range PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth*pt.Y
		self.los[idx] = LosUnknown
	}
}

func (self *World) MarkSeen(pos Pt2I) {
	if inTerrain(pos) {
		self.los[pos.X+pos.Y*mapWidth] = LosSeen
	}
}

func (self *World) GetLos(pos Pt2I) LosState {
	if inTerrain(pos) {
		return self.los[pos.X+pos.Y*mapWidth]
	}
	return LosUnknown
}

func (self *World) DoLos(center Pt2I) {
	const losRadius = 12

	blocks := func(vec Vec2I) bool { return self.BlocksSight(center.Plus(vec)) }

	outOfRadius := func(vec Vec2I) bool { return int(vec.Abs()) > losRadius }

	for pt := range LineOfSight(blocks, outOfRadius) {
		self.MarkSeen(center.Plus(pt))
	}
}

func (self *World) BlocksSight(pos Pt2I) bool {
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
	graph := NewSparseMatrixGraph()
	area.FindConnectingWalls(graph)
	doors := DoorLocations(graph)

	for pt := range PtIter(0, 0, mapWidth, mapHeight) {
		x, y := pt.X, pt.Y
		if area.RoomAtPoint(x, y) != nil {
			self.SetTerrain(Pt2I{x, y}, TerrainFloor)
		} else {
			self.SetTerrain(Pt2I{x, y}, TerrainWall)
		}
	}

	for pt := range doors.Iter() {
		pt := pt.(Pt2I)
		// TODO: Convert bsp to use Pt2I
		self.SetTerrain(pt, TerrainDoor)
	}
}

func (self *World) makeCaveMap() {
	area := MakeCaveMap(mapWidth, mapHeight, 0.50)
	for pt := range PtIter(0, 0, mapWidth, mapHeight) {
		switch area[pt.X][pt.Y] {
		case CaveFloor:
			self.SetTerrain(pt, TerrainFloor)
		case CaveWall:
			self.SetTerrain(pt, TerrainWall)
		case CaveUnknown:
			self.SetTerrain(pt, TerrainWall)
		default:
			Die("Bad data %v in generated cave map.", area[pt.X][pt.Y])
		}
	}
}

func inTerrain(pos Pt2I) bool {
	return pos.X >= 0 && pos.Y >= 0 && pos.X < mapWidth && pos.Y < mapHeight
}

func (self *World) GetTerrain(pos Pt2I) TerrainType {
	if inTerrain(pos) {
		return self.terrain[pos.X+pos.Y*mapWidth]
	}
	return TerrainIndeterminate
}

func (self *World) SetTerrain(pos Pt2I, t TerrainType) {
	if inTerrain(pos) {
		self.terrain[pos.X+pos.Y*mapWidth] = t
	}
}

func (self *World) EntitiesAt(pos Pt2I) <-chan Entity {
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


func (self *World) IsOpen(pos Pt2I) bool {
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

func (self *World) GetSpawnPos() (pos Pt2I) {
	pos, ok := self.GetMatchingPos(
		func(pos Pt2I) bool { return self.isSpawnPos(pos) })
	// XXX: Maybe this shouldn't be an assert, since a situation where no
	// spawn pos can be found can occur during play.
	Assert(ok, "Couldn't find open spawn position.")
	return
}

func (self *World) isSpawnPos(pos Pt2I) bool {
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

func (self *World) GetMatchingPos(f func(Pt2I) bool) (pos Pt2I, found bool) {
	const tries = 1024

	for i := 0; i < tries; i++ {
		x, y := rand.Intn(mapWidth), rand.Intn(mapHeight)
		pos = Pt2I{x, y}
		if f(pos) {
			return pos, true
		}
	}

	// RNG has failed us, let's do an exhaustive search...
	for pt := range PtIter(0, 0, mapWidth, mapHeight) {
		if f(pt) {
			return pt, true
		}
	}

	// There really doesn't seem to be any open positions.
	return Pt2I{0, 0}, false
}


func (self *World) drawTerrain() {
	for pt := range PtIter(0, 0, mapWidth, mapHeight) {
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
	PredicateSort(entityEarlierInDrawOrder, seq)

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
