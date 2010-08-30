package teratogen

import (
	"compress/gzip"
	"exp/iterable"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"hyades/mem"
	"io"
	"os"
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
	GetManager().Handler(FovComponent).Add(globals.AreaId, NewFov())

	endDepth := 15

	switch {
	case depth <= 3:
		GetArea().MakeBSPMap()
	case depth <= 6:
		GetArea().MakeCellarMap()
	case depth <= 9:
		GetArea().MakeCaveMap()
	case depth <= 12:
		GetArea().MakeRuinsMap()
	case depth <= 15:
		GetArea().MakeVisceraMap()
	default:
		depth = 15
		GetArea().MakeVisceraMap()
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
	GetFov().DoFov(GetPos(playerId))

	spawnEntities(depth)

	globals.CurrentLevel = int32(depth)
}

func spawnEntities(depth int) {
	spawns := makeSpawnDistribution(depth)
	for i := 0; i < spawnsPerLevel; i++ {
		proto := spawns.Sample(rand.Float64()).(string)
		SpawnRandomPos(proto)
	}
}

func Destroy(id entity.Id) {
	SetParent(id, entity.NilId)
	if id == PlayerId() {
		if /*GameRunning() */ false {
			// Ensure gameover if player is destroyed by unknown means.
			GameOver("were wiped out of existence.")
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
	// NOTE: The (*Type)(nil) pattern is for passing type information to a
	// reflection-using function without using any values.
	result = entity.NewManager()
	result.SetHandler(GlobalsComponent, new(Globals))
	result.SetHandler(AreaComponent, entity.NewContainer((*Area)(nil)))
	result.SetHandler(FovComponent, entity.NewContainer((*Fov)(nil)))
	result.SetHandler(PosComponent, entity.NewContainer((*Position)(nil)))
	result.SetHandler(NameComponent, entity.NewContainer((*Name)(nil)))
	result.SetHandler(CreatureComponent, entity.NewContainer((*Creature)(nil)))
	result.SetHandler(ItemComponent, entity.NewContainer((*Item)(nil)))
	result.SetHandler(DecalComponent, entity.NewContainer((*Decal)(nil)))
	result.SetHandler(MutationsComponent, entity.NewContainer((*Mutations)(nil)))
	result.SetHandler(WeaponComponent, entity.NewContainer((*Weapon)(nil)))
	result.SetHandler(FixedInventoryComponent, entity.NewContainer((*FixedInventory)(nil)))

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

func GetFov() *Fov {
	return GetManager().Handler(FovComponent).Get(GetGlobals().AreaId).(*Fov)
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

// CreatureAt returns the id of a creature at a given position, if there is
// one. It's assumed that only one creature can be present at one position at
// once.
func CreatureAt(pos geom.Pt2I) entity.Id {
	for o := range iterable.Filter(EntitiesAt(pos), EntityFilterFn(IsCreature)).Iter() {
		return o.(entity.Id)
	}
	return entity.NilId
}

func Creatures() iterable.Iterable {
	return iterable.Filter(Entities(), EntityFilterFn(IsCreature))
}

func OtherCreatures(excludedId interface{}) iterable.Iterable {
	pred := func(o interface{}) bool { return o != excludedId && IsCreature(o.(entity.Id)) }
	return iterable.Filter(Entities(), pred)
}

const saveMagicId = "SAVE"
const saveVersion = 0

// SaveGame saves the game state to a file. The file gets a header:
//   0 - 3: Magic bytes "SAVE" to identify the savefile
//   4 - 5: Save version as little-endian int16
//   6: 0 if the save isn't compressed. 1 if it's compressed with gzip.
//   7 -: World save data.
func SaveGame(fileName string, useGzip bool) (err os.Error) {
	var saveFile io.WriteCloser

	saveFile, err = os.Open(fileName, os.O_WRONLY|os.O_CREAT, 0666)
	if err != nil {
		return err
	}

	// Write the save header
	saveFile.Write([]byte(saveMagicId))

	mem.WriteFixed(saveFile, int16(saveVersion))

	// XXX: Awkward bool -> byte conversion, since we don't have ternary ops.
	if useGzip {
		mem.WriteFixed(saveFile, byte(1))
	} else {
		mem.WriteFixed(saveFile, byte(0))
	}

	if useGzip {
		saveFile, _ = gzip.NewWriter(saveFile)
	}

	// Write the actual save data.
	GetContext().Serialize(saveFile)

	saveFile.Close()
	return
}

func LoadGame(fileName string) (err os.Error) {
	var loadFile io.ReadCloser

	loadFile, err = os.Open(fileName, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}

	// Load the header.
	fileMagic := make([]byte, 4)
	loadFile.Read(fileMagic)
	if string(fileMagic) != saveMagicId {
		return os.NewError("Not a valid savefile.")
	}

	version := mem.ReadInt16(loadFile)
	if !isValidSaveVersion(version) {
		return os.NewError("Incompatible save version.")
	}

	useGzip := mem.ReadByte(loadFile)
	if useGzip != 0 {
		// If the save is compressed, switch to a gzip reader.
		loadFile, _ = gzip.NewReader(loadFile)
	}

	// Load the actual data.
	GetContext().Deserialize(loadFile)

	loadFile.Close()
	return
}

func isValidSaveVersion(version int16) bool {
	// Make the logic here more complex if we start getting backwards
	// incompatible too. For now, the basic assumption is that future save
	// versions can't be used.
	return version <= saveVersion
}
