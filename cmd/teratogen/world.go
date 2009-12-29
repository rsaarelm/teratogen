package main

import (
	"container/vector"
	"fmt"
	"hyades/alg"
	"hyades/dbg"
	"hyades/geom"
	"hyades/mem"
	"hyades/num"
	"io"
	"math"
	"rand"
	"reflect"
)

const mapWidth = 40
const mapHeight = 20

const spawnsPerLevel = 32

const numTerrainCells = mapWidth * mapHeight

var world *World

func DrawPos(pos geom.Pt2I) (screenX, screenY int) {
	return TileW*pos.X + xDrawOffset, TileH*pos.Y + yDrawOffset
}

func CenterDrawPos(pos geom.Pt2I) (screenX, screenY int) {
	return TileW*pos.X + xDrawOffset + TileW/2, TileH*pos.Y + yDrawOffset + TileH/2
}

func Draw(spriteId string, x, y int) {
	sx, sy := DrawPos(geom.Pt2I{x, y})
	DrawSprite(spriteId, sx, sy)
}

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

// Put item classes before creature classes, so we can use this to control
// draw order as well.
type EntityClass int

// XXX: Save compatibility is easily broken with this as adding things in the
// categories displaces the values further up.

const (
	EmptyEntityClass EntityClass = iota

	// Item classes
	GlobeEntityClass // Globe items are used when stepped on.
	ItemEntityClass  // Items can be picked up and dropped.

	// Creature classes
	CreatureEntityClassStartMarker

	PlayerEntityClass
	EnemyEntityClass
)

type entityPrototype struct {
	Name     string
	Parent   string
	IconId   string
	Class    EntityClass
	Scarcity int
	MinDepth int
	Props    map[string]interface{}
}

func NewPrototype(name, parent, iconId string, class EntityClass, scarcity, minDepth int, props ...) (result *entityPrototype) {
	result = new(entityPrototype)
	result.Name = name
	result.Parent = parent
	result.IconId = iconId
	result.Class = class
	result.Scarcity = scarcity
	result.MinDepth = minDepth

	// Custom settings from varargs.
	v := reflect.NewValue(props).(*reflect.StructValue)
	dbg.Assert(v.NumField()%2 == 0, "NewPrototype: Proplist length is odd.")
	result.Props = make(map[string]interface{})
	for i := 0; i < v.NumField(); i += 2 {
		result.Props[v.Field(i).Interface().(string)] = v.Field(i + 1).Interface()
	}
	return
}

func (self *entityPrototype) applyProps(prototypes map[string]*entityPrototype, target *Entity) {
	if parent, ok := prototypes[self.Parent]; ok {
		parent.applyProps(prototypes, target)
	}
	for key, val := range self.Props {
		target.Set(key, val)
	}
}

func (self *entityPrototype) MakeEntity(prototypes map[string]*entityPrototype, target *Entity) {
	target.IconId = self.IconId
	target.Name = self.Name
	target.Class = self.Class
	self.applyProps(prototypes, target)
}

func (self *entityPrototype) SpawnWeight(depth int) (result float64) {
	const epsilon = 1e-7
	const outOfDepthFactor = 2.0
	scarcity := float64(self.Scarcity)

	if depth < self.MinDepth {
		// Exponentially increase the scarcity for each level out of depth
		outOfDepth := self.MinDepth - depth
		scarcity *= math.Pow(outOfDepthFactor, float64(outOfDepth))
	}

	result = 1.0 / scarcity
	// Make too scarse weights just plain zero.
	if result < epsilon {
		result = 0.0
	}
	return
}

var prototypes = map[string]*entityPrototype{
	// Base prototype for creatures.
	"creature": NewPrototype("creature", "", "", EnemyEntityClass, -1, 0,
		FlagObstacle, 1,
		PropStrength, Fair,
		PropToughness, Fair,
		PropMeleeSkill, Fair,
		PropScale, 0,
		PropWounds, 0,
		PropDensity, 0),
	"protagonist": NewPrototype("protagonist", "creature", "chars:0", PlayerEntityClass, -1, 0,
		PropStrength, Superb,
		PropToughness, Good,
		PropMeleeSkill, Good),
	"zombie": NewPrototype("zombie", "creature", "chars:1", EnemyEntityClass, 100, 0,
		PropStrength, Fair,
		PropToughness, Poor,
		PropMeleeSkill, Fair),
	"dogthing": NewPrototype("dog-thing", "creature", "chars:2", EnemyEntityClass, 150, 0,
		PropStrength, Fair,
		PropToughness, Fair,
		PropMeleeSkill, Good,
		PropScale, -1),
	"ogre": NewPrototype("ogre", "creature", "chars:15", EnemyEntityClass, 600, 5,
		PropStrength, Great,
		PropToughness, Great,
		PropMeleeSkill, Fair,
		PropScale, 3),
	"boss1": NewPrototype("elder spawn", "creature", "chars:5", EnemyEntityClass, 3000, 10,
		PropStrength, Legendary,
		PropToughness, Legendary,
		PropMeleeSkill, Superb,
		PropScale, 5),
	"globe": NewPrototype("health globe", "", "items:1", GlobeEntityClass, 30, 0),
	"plantpot": NewPrototype("plant pot", "", "items:3", ItemEntityClass, 200, 0),
	"pistol": NewPrototype("pistol", "", "items:4", ItemEntityClass, 200, 0),
}

type LosState byte

const (
	LosUnknown LosState = iota
	LosMapped
	LosSeen
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

func IsObstacleTerrain(terrain TerrainType) bool {
	switch terrain {
	case TerrainWall, TerrainDirt:
		return true
	}
	return false
}

// Skinning data for a terrain tile set, describes the outward appearance of a
// type of terrain.
type TerrainTile struct {
	IconId string
	Name   string
}


type Guid string


type World struct {
	playerId     Guid
	entities     map[Guid]*Entity
	terrain      []TerrainType
	los          []LosState
	guidCounter  uint64
	currentLevel int32
}

func NewWorld() (result *World) {
	result = new(World)
	world = result
	result.entities = make(map[Guid]*Entity)
	result.initTerrain()

	player := result.Spawn(prototypes["protagonist"])
	result.playerId = player.GetGuid()

	result.InitLevel(1)

	return
}

func SetWorld(newWorld *World) { world = newWorld }

func GetWorld() *World {
	dbg.AssertNotNil(world, "World not initialized.")
	return world
}

func (self *World) Draw() {
	self.drawTerrain()
	self.drawEntities()
}

func (self *World) GetPlayer() *Entity { return self.entities[self.playerId] }

func (self *World) GetEntity(guid Guid) *Entity {
	if guid == *new(Guid) {
		return nil
	}
	ent, ok := self.entities[guid]
	dbg.Assert(ok, "GetEntity: Entity '%v' not found", guid)
	return ent
}

func (self *World) DestroyEntity(ent *Entity) {
	if ent == self.GetPlayer() {
		GameOver("was wiped out of existence.")
		return
	}
	self.entities[ent.GetGuid()] = ent, false
}

func (self *World) Spawn(prototype *entityPrototype) *Entity {
	guid := self.getGuid("")
	ent := NewEntity(guid)
	prototype.MakeEntity(prototypes, ent)
	self.entities[guid] = ent
	return ent
}

func (self *World) SpawnAt(prototype *entityPrototype, pos geom.Pt2I) (result *Entity) {
	result = self.Spawn(prototype)
	result.MoveAbs(pos)
	return
}

func (self *World) SpawnRandomPos(prototype *entityPrototype) (result *Entity) {
	return self.SpawnAt(prototype, self.GetSpawnPos())
}

func (self *World) InitLevel(depth int) {
	// Keep the player around even though the other entities get munged.
	// TODO: When we start having inventories, keep the player's items too.
	player := self.GetPlayer()

	self.currentLevel = int32(depth)

	self.initTerrain()

	// Bring over player object and player's inventory.
	keep := new(vector.Vector)
	keep.Push(player)
	self.entities[self.playerId] = player
	for ent := range player.RecursiveContents() {
		keep.Push(ent)
	}

	self.entities = make(map[Guid]*Entity)

	for i := range keep.Iter() {
		ent := i.(*Entity)
		self.entities[ent.GetGuid()] = ent
	}

	if num.WithProb(0.5) {
		self.makeCaveMap()
	} else {
		self.makeBSPMap()
	}

	self.SetTerrain(self.GetSpawnPos(), TerrainStairDown)

	player.MoveAbs(self.GetSpawnPos())
	self.DoLos(player.GetPos())

	spawns := makeSpawnDistribution(depth)
	for i := 0; i < spawnsPerLevel; i++ {
		proto := spawns.Sample(rand.Float64()).(*entityPrototype)
		// XXX: Get this wrapped in a func.
		guid := self.getGuid("")
		ent := NewEntity(guid)
		self.entities[guid] = ent

		proto.MakeEntity(prototypes, ent)
		ent.MoveAbs(self.GetSpawnPos())
	}
}

func makeSpawnDistribution(depth int) num.WeightedDist {
	weightFn := func(item interface{}) float64 { return item.(*entityPrototype).SpawnWeight(depth) }
	values := make([]interface{}, len(prototypes))
	i := 0
	for _, val := range prototypes {
		values[i] = val
		i++
	}
	return num.MakeWeightedDist(weightFn, values)
}

func (self *World) CurrentLevelNum() int { return int(self.currentLevel) }

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
			self.SetTerrain(pt, TerrainDirt)
		case CaveUnknown:
			self.SetTerrain(pt, TerrainDirt)
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

func (self *World) EntitiesAt(pos geom.Pt2I) <-chan *Entity {
	c := make(chan *Entity)
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

func (self *World) IterEntities() <-chan *Entity {
	c := make(chan *Entity)
	go func() {
		for _, ent := range self.entities {
			c <- ent
		}
		close(c)
	}()
	return c
}

func (self *World) IterCreatures() <-chan *Entity {
	c := make(chan *Entity)
	go func() {
		for _, ent := range self.entities {
			if IsCreature(ent) {
				c <- ent
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
		if e.Has(FlagObstacle) {
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
		idx := self.GetTerrain(pt)
		front := self.GetTerrain(pt.Plus(geom.Vec2I{0, 1}))
		// XXX: Hack to get the front tile visuals
		if idx == TerrainWall && front != TerrainWall && front != TerrainDoor {
			idx = TerrainWallFront
		}
		if idx == TerrainDirt && front != TerrainDirt && front != TerrainDoor {
			idx = TerrainDirtFront
		}
		Draw(tileset1[idx], pt.X, pt.Y)
	}
}

func (self *World) drawEntities() {
	// Make a vector of the entities sorted in draw order.
	seq := new(vector.Vector)
	for e := range self.IterEntities() {
		if e.GetParent() != nil {
			continue
		}
		seq.Push(e)
	}
	alg.PredicateSort(entityEarlierInDrawOrder, seq)

	for sorted := range seq.Iter() {
		e := sorted.(*Entity)
		pos := e.GetPos()
		seen := self.GetLos(pos) == LosSeen
		mapped := seen || self.GetLos(pos) == LosMapped
		// TODO: Draw static (item) entities from map memory.
		if mapped {
			if seen || !IsMobile(e) {
				Draw(e.IconId, pos.X, pos.Y)
			}
		}
	}
}

func entityEarlierInDrawOrder(i, j interface{}) bool {
	return i.(*Entity).GetClass() < j.(*Entity).GetClass()
}

func (self *World) getGuid(name string) (result Guid) {
	// If the guid's already in use, keep incrementing the counter until we get a fresh one.
	for ok := true; ok; _, ok = self.entities[result] {
		result = Guid(fmt.Sprintf("%v#%v", name, self.guidCounter))
		self.guidCounter++
	}

	return
}

func (self *World) Serialize(out io.Writer) {
	mem.WriteString(out, string(self.playerId))
	mem.WriteInt64(out, int64(self.guidCounter))
	mem.WriteInt32(out, self.currentLevel)

	mem.WriteNTimes(out, len(self.terrain), func(i int, out io.Writer) { mem.WriteByte(out, byte(self.terrain[i])) })
	mem.WriteNTimes(out, len(self.los), func(i int, out io.Writer) { mem.WriteByte(out, byte(self.los[i])) })

	mem.WriteInt32(out, int32(len(self.entities)))
	for guid, ent := range self.entities {
		// Write the guid
		mem.WriteString(out, string(guid))
		// Then gob-save the entity using the factory.
		ent.Serialize(out)
	}
}

func (self *World) Deserialize(in io.Reader) {
	self.playerId = Guid(mem.ReadString(in))
	self.guidCounter = uint64(mem.ReadInt64(in))
	self.currentLevel = mem.ReadInt32(in)

	mem.ReadNTimes(in,
		func(count int) { self.terrain = make([]TerrainType, count) },
		func(i int, in io.Reader) { self.terrain[i] = TerrainType(mem.ReadByte(in)) })
	mem.ReadNTimes(in,
		func(count int) { self.los = make([]LosState, count) },
		func(i int, in io.Reader) { self.los[i] = LosState(mem.ReadByte(in)) })

	// TODO: Entities.
	self.entities = make(map[Guid]*Entity)
	for i, numEntities := 0, int(mem.ReadInt32(in)); i < numEntities; i++ {
		guid := Guid(mem.ReadString(in))

		ent := new(Entity)
		ent.Deserialize(in)
		self.entities[guid] = ent
	}
}
