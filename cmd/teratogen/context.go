package main

import (
	"hyades/dbg"
	"hyades/entity"
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

	world.areaId = self.manager.NewEntity()
	GetManager().Handler(AreaComponent).Add(world.areaId, NewArea())

	// TODO: Line-of-sight component
	world.initLos()

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
	world.DoLos(player.GetPos())

	spawns := makeSpawnDistribution(depth)
	for i := 0; i < spawnsPerLevel; i++ {
		proto := spawns.Sample(rand.Float64()).(string)
		ent := world.Spawn(proto)
		ent.MoveAbs(world.GetSpawnPos())
	}
}

func (self *Context) GetPlayer() *Blob { return self.getWorld().GetPlayer() }

func (self *Context) Deserialize(in io.Reader) {
	self.manager = makeManager()
	self.manager.Deserialize(in)
}

func (self *Context) Serialize(out io.Writer) { self.manager.Serialize(out) }

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

func (self *Context) getWorld() *World { return self.manager.Handler(WorldComponent).(*World) }

// XXX: Deprecated, try to work via context instead.
func GetWorld() *World {
	dbg.AssertNotNil(gContext, "World not initialized.")
	return GetContext().getWorld()
}

func GetArea() *Area {
	return GetManager().Handler(AreaComponent).Get(GetWorld().areaId).(*Area)
}

func GetBlobs() entity.Handler { return GetManager().Handler(BlobComponent) }
