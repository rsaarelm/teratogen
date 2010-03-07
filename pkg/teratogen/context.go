package teratogen

import (
	"exp/iterable"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"hyades/num"
	"io"
	"rand"
)

var gContext *Context

var gEffects Effects

// Context is the toplevel game content container
type Context struct {
	manager *entity.Manager
}

func InitEffects(fx Effects) { gEffects = fx }

func NewContext() (result *Context) {
	result = new(Context)
	result.manager = makeManager()
	gContext = result
	return
}

func Fx() Effects { return gEffects }

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

	playerId := Spawn("protagonist")
	globals.PlayerId = playerId

	self.EnterLevel(1)
}

func (self *Context) EnterLevel(depth int) {
	globals := GetGlobals()

	// Delete old area.
	self.manager.RemoveEntity(globals.AreaId)

	// Move player and inventory to the new level, ditch other entities.
	clearNonplayerEntities()

	// Make new area.
	globals.AreaId = self.manager.NewEntity()
	GetManager().Handler(AreaComponent).Add(globals.AreaId, NewArea())
	GetManager().Handler(LosComponent).Add(globals.AreaId, NewLos())

	endDepth := 20

	switch {
	case depth <= 4:
		GetArea().MakeBSPMap()
	case depth <= 8:
		GetArea().MakeCellarMap()
	case depth <= 12:
		GetArea().MakeCaveMap()
	case depth <= 16:
		GetArea().MakeRuinsMap()
	case depth <= 20:
		GetArea().MakeVisceraMap()
	default:
		if num.WithProb(0.5) {
			GetArea().MakeCaveMap()
		} else {
			GetArea().MakeBSPMap()
		}
	}

	if depth == endDepth {
		// End boss level. No stairs down. Has end boss.
		SpawnRandomPos("boss1")
	} else {
		if pos, ok := GetSpawnPos(); ok {
			GetArea().SetTerrain(pos, TerrainStairDown)
		} else {
			dbg.Die("Couldn't place stairs down.")
		}
	}

	playerId := PlayerId()
	if pos, ok := GetSpawnPos(); ok {
		SetPos(playerId, pos)
	} else {
		dbg.Die("Couldn't place player.")
	}
	GetLos().DoLos(GetPos(playerId))

	spawns := makeSpawnDistribution(depth)
	for i := 0; i < spawnsPerLevel; i++ {
		proto := spawns.Sample(rand.Float64()).(string)
		SpawnRandomPos(proto)
	}

	globals.CurrentLevel = int32(depth)
}

func Destroy(id entity.Id) {
	SetParent(id, entity.NilId)
	if id == PlayerId() {
		if /*GameRunning() */ false {
			// Ensure gameover if player is destroyed by unknown means.
			GameOver("was wiped out of existence.")
		}
		// XXX: The system can't currently handle the player entity being
		// removed.
		return
	}
	GetManager().RemoveEntity(id)
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
	result.SetHandler(PosComponent, entity.NewContainer(new(Position)))
	result.SetHandler(NameComponent, entity.NewContainer(new(Name)))
	result.SetHandler(CreatureComponent, entity.NewContainer(new(Creature)))
	result.SetHandler(ItemComponent, entity.NewContainer(new(Item)))

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

// Entities returns an iteration of all the entity ids in the game.
func Entities() iterable.Iterable { return GetManager().Entities() }

// EntitiesAt returns an iteration of the ids of positionable entities at a
// given position.
func EntitiesAt(pos geom.Pt2I) iterable.Iterable {
	posPred := func(obj interface{}) bool {
		id := obj.(entity.Id)
		return GetParent(id) == entity.NilId && HasPosComp(id) && GetPos(id).Equals(pos)
	}
	return iterable.Filter(Entities(), posPred)
}

func Creatures() iterable.Iterable {
	return iterable.Filter(Entities(), EntityFilterFn(IsCreature))
}

func OtherCreatures(excludedId interface{}) iterable.Iterable {
	pred := func(o interface{}) bool { return o != excludedId && IsCreature(o.(entity.Id)) }
	return iterable.Filter(Entities(), pred)
}
