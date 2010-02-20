package main

import (
	"exp/iterable"
	"hyades/entity"
	"hyades/geom"
	"hyades/num"
	"io"
	"rand"
)

var gContext *Context

// Context is the toplevel game content container
type Context struct {
	manager *entity.Manager
}

func NewContext() (result *Context) {
	result = new(Context)
	result.manager = makeManager()
	gContext = result
	return
}

func LoadContext(in io.Reader) (result *Context) {
	result = NewContext()
	result.Deserialize(in)
	return
}

// GetContext returns the global Context value.
func GetContext() (result *Context) { return gContext }

func GetManager() *entity.Manager { return GetContext().manager }

func (self *Context) InitGame() {
	globals := GetGlobals()

	player := Spawn("protagonist")
	globals.PlayerId = player.GetGuid()

	self.EnterLevel(1)
}

func (self *Context) EnterLevel(depth int) {
	globals := GetGlobals()

	// Delete old area.
	self.manager.RemoveEntity(globals.AreaId)

	// Make new area.
	globals.AreaId = self.manager.NewEntity()
	GetManager().Handler(AreaComponent).Add(globals.AreaId, NewArea())
	GetManager().Handler(LosComponent).Add(globals.AreaId, NewLos())

	// Move player and inventory to the new level, ditch other entities.
	clearNonplayerEntities()

	if num.WithProb(0.5) {
		GetArea().MakeCaveMap()
	} else {
		GetArea().MakeBSPMap()
	}

	GetArea().SetTerrain(GetSpawnPos(), TerrainStairDown)

	playerId := PlayerId()
	PosComp(playerId).MoveAbs(GetSpawnPos())
	GetLos().DoLos(GetPos(playerId))

	spawns := makeSpawnDistribution(depth)
	for i := 0; i < spawnsPerLevel; i++ {
		proto := spawns.Sample(rand.Float64()).(string)
		SpawnRandomPos(proto)
	}

	globals.CurrentLevel = int32(depth)
}

func GetBlob(guid entity.Id) *Blob { return GetBlobs().Get(guid).(*Blob) }

func DestroyBlob(ent *Blob) {
	ent.RemoveSelf()
	if ent.GetGuid() == PlayerId() {
		if /*GameRunning() */ false {
			// Ensure gameover if player is destroyed by unknown means.
			GameOver("was wiped out of existence.")
		}
		// XXX: The system can't currently handle the player entity being
		// removed.
		return
	}
	GetManager().RemoveEntity(ent.GetGuid())
}

func PlayerId() entity.Id { return GetGlobals().PlayerId }

func (self *Context) Deserialize(in io.Reader) {
	self.manager = makeManager()
	self.manager.Deserialize(in)
}

func (self *Context) Serialize(out io.Writer) { self.manager.Serialize(out) }

func makeManager() (result *entity.Manager) {
	result = entity.NewManager()
	result.SetHandler(GlobalsComponent, new(Globals))
	result.SetHandler(AreaComponent, entity.NewContainer(new(Area)))
	result.SetHandler(LosComponent, entity.NewContainer(new(Los)))
	result.SetHandler(BlobComponent, entity.NewContainer(new(Blob)))
	result.SetHandler(PosComponent, entity.NewContainer(new(Position)))
	result.SetHandler(ContainComponent, entity.NewRelation(entity.OneToMany))
	result.SetHandler(MeleeEquipComponent, entity.NewRelation(entity.OneToOne))
	result.SetHandler(GunEquipComponent, entity.NewRelation(entity.OneToOne))
	result.SetHandler(ArmorEquipComponent, entity.NewRelation(entity.OneToOne))

	return
}

func GetGlobals() *Globals { return GetManager().Handler(GlobalsComponent).(*Globals) }

func GetArea() *Area {
	return GetManager().Handler(AreaComponent).Get(GetGlobals().AreaId).(*Area)
}

func GetLos() *Los {
	return GetManager().Handler(LosComponent).Get(GetGlobals().AreaId).(*Los)
}

func GetBlobs() entity.Handler { return GetManager().Handler(BlobComponent) }

func Entities() iterable.Iterable {
	return iterable.Map(GetBlobs().EntityComponents(), entity.IdComponent2Component)
}

func EntitiesAt(pos geom.Pt2I) iterable.Iterable {
	posPred := func(obj interface{}) bool {
		e := obj.(*Blob)
		return e.GetParent() == nil && e.GetPos().Equals(pos)
	}
	return iterable.Filter(Entities(), posPred)
}

func Creatures() iterable.Iterable {
	return iterable.Filter(Entities(), func(o interface{}) bool { return IsCreature(o.(*Blob).GetGuid()) })
}

func OtherCreatures(excluded interface{}) iterable.Iterable {
	pred := func(o interface{}) bool { return o != excluded && IsCreature(o.(*Blob).GetGuid()) }
	return iterable.Filter(Entities(), pred)
}
