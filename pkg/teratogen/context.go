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
	// TODO: Ditch World

	// TODO: "Globals" component to hold player id

	world := self.getWorld()

	player := world.Spawn("protagonist")
	world.playerId = player.GetGuid()

	self.EnterLevel(1)
}

func (self *Context) EnterLevel(depth int) {
	world := self.getWorld()

	// Delete old area.
	self.manager.RemoveEntity(world.areaId)

	// Make new area.
	world.areaId = self.manager.NewEntity()
	GetManager().Handler(AreaComponent).Add(world.areaId, NewArea())
	GetManager().Handler(LosComponent).Add(world.areaId, NewLos())

	// Move player and inventory to the new level, ditch other entities.
	world.clearNonplayerEntities()

	if num.WithProb(0.5) {
		GetArea().MakeCaveMap()
	} else {
		GetArea().MakeBSPMap()
	}

	GetArea().SetTerrain(world.GetSpawnPos(), TerrainStairDown)

	player := world.GetPlayer()
	player.MoveAbs(world.GetSpawnPos())
	GetLos().DoLos(player.GetPos())

	spawns := makeSpawnDistribution(depth)
	for i := 0; i < spawnsPerLevel; i++ {
		proto := spawns.Sample(rand.Float64()).(string)
		ent := world.Spawn(proto)
		ent.MoveAbs(world.GetSpawnPos())
	}
}

func GetPlayer() *Blob { return GetWorld().GetPlayer() }

func (self *Context) Deserialize(in io.Reader) {
	self.manager = makeManager()
	self.manager.Deserialize(in)
}

func (self *Context) Serialize(out io.Writer) { self.manager.Serialize(out) }


func makeManager() (result *entity.Manager) {
	result = entity.NewManager()
	result.SetHandler(WorldComponent, new(World))
	result.SetHandler(AreaComponent, entity.NewContainer(new(Area)))
	result.SetHandler(LosComponent, entity.NewContainer(new(Los)))
	result.SetHandler(BlobComponent, entity.NewContainer(new(Blob)))

	result.SetHandler(ContainComponent, entity.NewRelation(entity.OneToMany))
	result.SetHandler(MeleeEquipComponent, entity.NewRelation(entity.OneToOne))
	result.SetHandler(GunEquipComponent, entity.NewRelation(entity.OneToOne))
	result.SetHandler(ArmorEquipComponent, entity.NewRelation(entity.OneToOne))

	return
}

func (self *Context) getWorld() *World { return self.manager.Handler(WorldComponent).(*World) }

// XXX: Deprecated, try to work via context instead.
func GetWorld() *World {
	dbg.AssertNotNil(gContext, "World not initialized.")
	return GetContext().getWorld()
}

func GetArea() *Area {
	return GetManager().Handler(AreaComponent).Get(GetWorld().areaId).(*Area)
}

func GetLos() *Los { return GetManager().Handler(LosComponent).Get(GetWorld().areaId).(*Los) }

func GetBlobs() entity.Handler { return GetManager().Handler(BlobComponent) }

// Entities iterates through all the game objects in the current context.
// XXX: Currently only iterates entities with a blob component.
func Entities() iterable.Iterable {
	return iterable.Map(GetBlobs().EntityComponents(), entity.IdComponent2Component)
}

// EntitiesAt iterates through all entities which have a positional component
// and are located at a given point.
func EntitiesAt(pos geom.Pt2I) iterable.Iterable {
	posPred := func(obj interface{}) bool {
		e := obj.(*Blob)
		return e.GetParent() == nil && e.GetPos().Equals(pos)
	}
	return iterable.Filter(Entities(), posPred)
}
