package main

import (
	"container/vector"
	"exp/iterable"
	"hyades/alg"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"hyades/gfx"
	"hyades/mem"
	"hyades/num"
	"io"
	"rand"
)


const spawnsPerLevel = 32

var manager *entity.Manager

const WorldComponent = entity.ComponentFamily("world")

func DrawPos(pos geom.Pt2I) (screenX, screenY int) {
	return TileW*pos.X + xDrawOffset, TileH*pos.Y + yDrawOffset
}

func CenterDrawPos(pos geom.Pt2I) (screenX, screenY int) {
	return TileW*pos.X + xDrawOffset + TileW/2, TileH*pos.Y + yDrawOffset + TileH/2
}

func Draw(g gfx.Graphics, spriteId string, x, y int) {
	sx, sy := DrawPos(geom.Pt2I{x, y})
	DrawSprite(g, spriteId, sx, sy)
}

type LosState byte

const (
	LosUnknown LosState = iota
	LosMapped
	LosSeen
)

func GetManager() *entity.Manager { return manager }

type World struct {
	playerId     entity.Id
	areaId       entity.Id
	los          []LosState
	currentLevel int32
}

func makeManager() (result *entity.Manager) {
	result = entity.NewManager()
	result.SetHandler(WorldComponent, new(World))
	result.SetHandler(AreaComponent, entity.NewContainer(new(Area)))
	result.SetHandler(BlobComponent, entity.NewContainer(new(Blob)))

	result.SetHandler(ContainComponent, entity.NewRelation(entity.OneToMany))
	result.SetHandler(MeleeEquipComponent, entity.NewRelation(entity.OneToOne))
	result.SetHandler(GunEquipComponent, entity.NewRelation(entity.OneToOne))
	result.SetHandler(ArmorEquipComponent, entity.NewRelation(entity.OneToOne))

	return
}

func LoadGame(in io.Reader) {
	manager = makeManager()
	manager.Deserialize(in)
}

func SaveGame(out io.Writer) { manager.Serialize(out) }

func InitWorld() {
	manager = makeManager()
	world := GetWorld()
	areas := manager.Handler(AreaComponent).(*entity.Container)

	world.areaId = manager.NewEntity()
	area := NewArea()
	areas.Add(world.areaId, area)

	world.initLos()

	player := world.Spawn(prototypes["protagonist"])
	world.playerId = player.GetGuid()

	return
}

func GetWorld() *World {
	dbg.AssertNotNil(manager, "World not initialized.")
	return GetManager().Handler(WorldComponent).(*World)
}

func GetArea() *Area {
	return GetManager().Handler(AreaComponent).Get(GetWorld().areaId).(*Area)
}

func GetBlobs() entity.Handler { return GetManager().Handler(BlobComponent) }

func (self *World) Draw(g gfx.Graphics) {
	self.drawTerrain(g)
	self.drawEntities(g)
}

func (self *World) GetPlayer() *Blob { return self.GetEntity(self.playerId) }

func (self *World) GetEntity(guid entity.Id) *Blob {
	if guid == *new(entity.Id) {
		return nil
	}
	return GetBlobs().Get(guid).(*Blob)
}

func (self *World) DestroyEntity(ent *Blob) {
	ent.RemoveSelf()
	if ent == self.GetPlayer() {
		if GameRunning() {
			// Ensure gameover if player is destroyed by unknown means.
			GameOver("was wiped out of existence.")
		}
		// XXX: The system can't currently handle the player entity being
		// removed.
		return
	}
	GetManager().RemoveEntity(ent.GetGuid())
}

func (self *World) Spawn(prototype *entityPrototype) *Blob {
	guid := GetManager().NewEntity()
	blob := NewEntity(guid)

	GetBlobs().Add(guid, blob)

	prototype.MakeEntity(prototypes, blob)
	return blob
}

func (self *World) SpawnAt(prototype *entityPrototype, pos geom.Pt2I) (result *Blob) {
	result = self.Spawn(prototype)
	result.MoveAbs(pos)
	return
}

func (self *World) SpawnRandomPos(prototype *entityPrototype) (result *Blob) {
	return self.SpawnAt(prototype, self.GetSpawnPos())
}

func (self *World) clearNonplayerEntities() {
	// Bring over player object and player's inventory.
	player := self.GetPlayer()
	keep := make(map[entity.Id]bool)
	keep[player.GetGuid()] = true
	for ent := range player.RecursiveContents().Iter() {
		keep[ent.(*Blob).GetGuid()] = true
	}

	for o := range GetBlobs().EntityComponents().Iter() {
		pair := o.(*entity.IdComponent)
		if _, ok := keep[pair.Entity]; !ok {
			defer GetManager().RemoveEntity(pair.Entity)
		}
	}
}

func (self *World) InitLevel(depth int) {
	self.currentLevel = int32(depth)

	self.initLos()

	self.clearNonplayerEntities()

	if num.WithProb(0.5) {
		GetArea().MakeCaveMap()
	} else {
		GetArea().MakeBSPMap()
	}

	GetArea().SetTerrain(self.GetSpawnPos(), TerrainStairDown)

	player := self.GetPlayer()
	player.MoveAbs(self.GetSpawnPos())
	self.DoLos(player.GetPos())

	spawns := makeSpawnDistribution(depth)
	for i := 0; i < spawnsPerLevel; i++ {
		proto := spawns.Sample(rand.Float64()).(*entityPrototype)
		ent := self.Spawn(proto)
		ent.MoveAbs(self.GetSpawnPos())
	}
}

func makeSpawnDistribution(depth int) num.WeightedDist {
	weightFn := func(item interface{}) float64 {
		proto := item.(*entityPrototype)
		return SpawnWeight(proto.Scarcity, proto.MinDepth, depth)
	}
	values := make([]interface{}, len(prototypes))
	i := 0
	for _, val := range prototypes {
		values[i] = val
		i++
	}
	return num.MakeWeightedDist(weightFn, values)
}

func (self *World) CurrentLevelNum() int { return int(self.currentLevel) }

func (self *World) initLos() { self.los = make([]LosState, numTerrainCells) }

func (self *World) ClearLosSight() {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth*pt.Y
		if self.los[idx] == LosSeen {
			self.los[idx] = LosMapped
		}
	}
}

// Debug command that makes the entire map visible.
func (self *World) WizardEye() {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth*pt.Y
		self.los[idx] = LosSeen
	}
}

func (self *World) ClearLosMapped() {
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth*pt.Y
		self.los[idx] = LosUnknown
	}
}

func (self *World) MarkSeen(pos geom.Pt2I) {
	if GetArea().InArea(pos) {
		self.los[pos.X+pos.Y*mapWidth] = LosSeen
	}
}

func (self *World) GetLos(pos geom.Pt2I) LosState {
	if GetArea().InArea(pos) {
		return self.los[pos.X+pos.Y*mapWidth]
	}
	return LosUnknown
}

func (self *World) DoLos(center geom.Pt2I) {
	const losRadius = 12

	blocks := func(vec geom.Vec2I) bool { return GetArea().BlocksSight(center.Plus(vec)) }

	outOfRadius := func(vec geom.Vec2I) bool { return int(vec.Abs()) > losRadius }

	for pt := range geom.LineOfSight(blocks, outOfRadius) {
		self.MarkSeen(center.Plus(pt))
	}
}

func (self *World) Entities() iterable.Iterable {
	return iterable.Map(GetBlobs().EntityComponents(), entity.IdComponent2Component)
}

func (self *World) EntitiesAt(pos geom.Pt2I) iterable.Iterable {
	posPred := func(obj interface{}) bool {
		e := obj.(*Blob)
		return e.GetParent() == nil && e.GetPos().Equals(pos)
	}
	return iterable.Filter(self.Entities(), posPred)
}

func (self *World) Creatures() iterable.Iterable {
	return iterable.Filter(self.Entities(), IsCreature)
}

func (self *World) OtherCreatures(excluded interface{}) iterable.Iterable {
	pred := func(o interface{}) bool { return o != excluded && IsCreature(o) }
	return iterable.Filter(self.Entities(), pred)
}

func (self *World) IsOpen(pos geom.Pt2I) bool {
	if IsObstacleTerrain(GetArea().GetTerrain(pos)) {
		return false
	}
	for o := range self.EntitiesAt(pos).Iter() {
		ent := o.(*Blob)
		if ent.Has(FlagObstacle) {
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
	if GetArea().GetTerrain(pos) == TerrainDoor {
		return false
	}
	if GetArea().GetTerrain(pos) == TerrainStairDown {
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

func (self *World) drawEntities(g gfx.Graphics) {
	// Make a vector of the entities sorted in draw order.
	seq := new(vector.Vector)
	for o := range self.Entities().Iter() {
		ent := o.(*Blob)
		if ent.GetParent() != nil {
			// Skip entities inside something.
			continue
		}
		seq.Push(ent)
	}
	alg.PredicateSort(entityEarlierInDrawOrder, seq)

	for sorted := range seq.Iter() {
		e := sorted.(*Blob)
		pos := e.GetPos()
		seen := self.GetLos(pos) == LosSeen
		mapped := seen || self.GetLos(pos) == LosMapped
		// TODO: Draw static (item) entities from map memory.
		if mapped {
			if seen || !IsMobile(e) {
				Draw(g, e.IconId, pos.X, pos.Y)
			}
		}
	}
}

func entityEarlierInDrawOrder(i, j interface{}) bool {
	return i.(*Blob).GetClass() < j.(*Blob).GetClass()
}

func (self *World) Serialize(out io.Writer) {
	mem.WriteFixed(out, int64(self.playerId))
	mem.WriteFixed(out, self.currentLevel)
	mem.WriteFixed(out, int64(self.areaId))

	mem.WriteNTimes(out, len(self.los), func(i int, out io.Writer) { mem.WriteFixed(out, byte(self.los[i])) })
}

func (self *World) Deserialize(in io.Reader) {
	self.playerId = entity.Id(mem.ReadInt64(in))
	self.currentLevel = mem.ReadInt32(in)
	self.areaId = entity.Id(mem.ReadInt64(in))

	mem.ReadNTimes(in,
		func(count int) { self.los = make([]LosState, count) },
		func(i int, in io.Reader) { self.los[i] = LosState(mem.ReadByte(in)) })
}

// Component handler interface stubs

func (self *World) Add(guid entity.Id, component interface{}) {
}

func (self *World) Remove(guid entity.Id) {}

func (self *World) Get(guid entity.Id) interface{} {
	return nil
}

func (self *World) EntityComponents() iterable.Iterable {
	return alg.EmptyIter()
}
