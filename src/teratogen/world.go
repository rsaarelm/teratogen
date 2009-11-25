package teratogen

import "fmt"
import "rand"
import "sync"

import . "fomalhaut"

const mapWidth = 80
const mapHeight = 40

const numTerrainCells = mapWidth * mapHeight

type Icon struct {
	IconId byte;
	Color RGB;
}

func (self *Icon)Draw(x, y int) {
	DrawCharRGB(x, y, int(self.IconId), self.Color);
}


// Behavioral terrain types.
type TerrainType byte const (
	// Used for terrain generation algorithms, set map to indeterminate
	// initially.
	TerrainIndeterminate = iota;
	TerrainWall;
	TerrainFloor;
	TerrainDoor;
)

type EntityType int const (
	EntityUnknown = iota;
	EntityPlayer;
	EntityZombie;
)

type LosState byte const (
	LosUnknown = iota;
	LosMapped;
	LosSeen;
)

var tileset1 = []Icon{
TerrainIndeterminate: Icon{'?', RGB{255, 0, 255}},
TerrainWall: Icon{'#', RGB{196, 64, 0}},
TerrainFloor: Icon{'.', RGB{196, 196, 196}},
TerrainDoor: Icon{'+', RGB{0, 196, 196}},
}

func IsObstacleTerrain(terrain TerrainType) bool {
	switch terrain {
	case TerrainWall:
		return true;
	}
	return false;
}

// Skinning data for a terrain tile set, describes the outward appearance of a
// type of terrain.
type TerrainTile struct {
	Icon;
	Name string;
}


type Drawable interface {
	Draw(x, y int);
}


type Guid string


type Entity interface {
	Drawable;
	// TODO: Entity-common stuff.
	IsObstacle() bool;
	GetPos() Pt2I;
	GetGuid() Guid;
	MoveAbs(pos Pt2I);
	Move(vec Vec2I);
}


type Creature struct {
	*Icon;
	guid Guid;
	Name string;
	pos Pt2I;
}

func (self *Creature) IsObstacle() bool { return true }

func (self *Creature) GetPos() Pt2I { return self.pos }

func (self *Creature) GetGuid() Guid { return self.guid }

// XXX: Assuming Pt2I to be a value type here.
func (self *Creature) MoveAbs(pos Pt2I) { self.pos = pos }

func (self *Creature) Move(vec Vec2I) { self.pos = self.pos.Plus(vec) }


type World struct {
	playerId Guid;
	Lock *sync.RWMutex;
	entities map[Guid] Entity;
	terrain []TerrainType;
	los []LosState;
	guidCounter uint64;
}

func NewWorld() (result *World) {
	result = new(World);
	result.entities = make(map[Guid] Entity);
	result.initTerrain();
	result.Lock = new(sync.RWMutex);

	result.playerId = Guid("player");
	player := result.Spawn(EntityPlayer);
	result.playerId = player.GetGuid();

	return;
}

func (self *World) Draw() {
	self.drawTerrain();
	self.drawEntities();
}

func (self *World) GetPlayer() *Creature {
	return self.entities[self.playerId].(*Creature);
}

func (self *World) Spawn(entityType EntityType) (result Entity) {
	guid := self.getGuid("");
	switch entityType {
	case EntityPlayer:
		result = &Creature{&Icon{'@', RGB{0xdd, 0xff, 0xff}}, guid, "protagonist", Pt2I{-1, -1}};
	case EntityZombie:
		result = &Creature{&Icon{'z', RGB{0x80, 0xa0, 0x80}}, guid, "zombie", Pt2I{-1, -1}};
	default:
		Die("Unknown entity type.");
	}
	self.entities[guid] = result;
	return;
}

func (self *World) SpawnAt(entityType EntityType, pos Pt2I) (result Entity) {
	result = self.Spawn(entityType);
	result.MoveAbs(pos);
	return;
}

func (self *World) SpawnRandomPos(entityType EntityType) (result Entity) {
	return self.SpawnAt(entityType, self.GetSpawnPos());
}

// TODO: Event system for changing world, event handler does lock/unlock, all
// changes in events. "Transactional database".

func (self *World) MovePlayer(vec Vec2I) {
	self.Lock.Lock();
	defer self.Lock.Unlock();

	player := self.GetPlayer();

	if self.IsOpen(player.GetPos().Plus(vec)) {
		player.Move(vec)
	}
}

func (self *World) InitLevel(num int) {
	// Keep the player around even though the other entities get munged.
	// TODO: When we start having inventories, keep the player's items too.
	player := self.GetPlayer();

	self.initTerrain();
	self.entities = make(map[Guid] Entity);
	self.entities[self.playerId] = player;

	self.makeBSPMap();

	player.MoveAbs(self.GetSpawnPos());

	for i := 0; i < 10 + num * 4; i++ {
		self.SpawnRandomPos(EntityZombie);
	}
}

func (self *World) initTerrain() {
	self.terrain = make([]TerrainType, numTerrainCells);
	self.los = make([]LosState, numTerrainCells);
}

func (self *World) ClearLosSight() {
	for pt := range PtIter(0, 0, mapWidth, mapHeight) {
		idx := pt.X + mapWidth * pt.Y;
		if self.los[idx] == LosSeen { self.los[idx] = LosMapped; }
	}
}

func (self *World) MarkSeen(pos Pt2I) {
	if inTerrain(pos) {
		self.los[pos.X + pos.Y * mapWidth] = LosSeen;
	}
}

func (self *World) GetLos(pos Pt2I) LosState {
	if inTerrain(pos) {
		return self.los[pos.X + pos.Y * mapWidth];
	}
	return LosUnknown;
}

func (self *World) DoLos(center Pt2I) {
	const losRadius = 8;

	blocks := func(vec Vec2I) bool {
		return self.BlocksSight(center.Plus(vec));
	};

	outOfRadius := func(vec Vec2I) bool {
		return int(vec.Abs()) > losRadius;
	};

	for pt := range LineOfSight(blocks, outOfRadius) {
		self.MarkSeen(center.Plus(pt));
	}
}

func (self *World) BlocksSight(pos Pt2I) bool {
	if IsObstacleTerrain(self.GetTerrain(pos)) {
		return true;
	}
	if self.GetTerrain(pos) == TerrainDoor { return true; }

	return false;
}

func (self *World) makeBSPMap() {
	area := MakeBspMap(1, 1, mapWidth - 2, mapHeight - 2);
	graph := NewSparseMatrixGraph();
	area.FindConnectingWalls(graph);
	doors := DoorLocations(graph);

	for pt := range PtIter(0, 0, mapWidth, mapHeight) {
		x, y := pt.X, pt.Y;
		if area.RoomAtPoint(x, y) != nil {
			self.SetTerrain(Pt2I{x, y}, TerrainFloor);
		} else {
			self.SetTerrain(Pt2I{x, y}, TerrainWall);
		}
	}

	for pt := range doors.Iter() {
		pt := pt.(Pt2I);
		// TODO: Convert bsp to use Pt2I
		self.SetTerrain(pt, TerrainDoor);
	}

}

func inTerrain(pos Pt2I) bool {
	return pos.X >= 0 && pos.Y >= 0 && pos.X < mapWidth && pos.Y < mapHeight;
}

func (self *World) GetTerrain(pos Pt2I) TerrainType {
	if inTerrain(pos) {
		return self.terrain[pos.X + pos.Y * mapWidth];
	}
	return TerrainIndeterminate;
}

func (self *World) SetTerrain(pos Pt2I, t TerrainType) {
	if inTerrain(pos) {
		self.terrain[pos.X + pos.Y * mapWidth] = t
	}
}

func (self *World) EntitiesAt(pos Pt2I) <-chan Entity {
	c := make(chan Entity);
	go func() {
		for _, ent := range self.entities {
			if ent.GetPos().Equals(pos) {
				c <- ent;
			}
		}
		close(c);
	}();
	return c;
}


func (self *World) IsOpen(pos Pt2I) bool {
	if IsObstacleTerrain(self.GetTerrain(pos)) {
		return false;
	}
	for e := range self.EntitiesAt(pos) {
		if e.IsObstacle() {
			return false;
		}
	}

	return true;
}

func (self *World) GetSpawnPos() (pos Pt2I) {
	pos, ok := self.GetMatchingPos(
		func(pos Pt2I) bool { return self.isSpawnPos(pos); });
	if !ok {
		Die("Couldn't find open position.");
	}
	return;
}

func (self *World) isSpawnPos(pos Pt2I) bool {
	if !self.IsOpen(pos) { return false; }
	if self.GetTerrain(pos) == TerrainDoor { return false; }
	return true;
}

func (self *World) GetMatchingPos(f func (Pt2I) bool) (pos Pt2I, found bool) {
	const tries = 1024;

	for i := 0; i < tries; i++ {
		x, y := rand.Intn(mapWidth), rand.Intn(mapHeight);
		pos = Pt2I{x, y};
		if f(pos) {
			return pos, true;
		}
	}

	// RNG has failed us, let's do an exhaustive search...
	for pt := range PtIter(0, 0, mapWidth, mapHeight) {
		if f(pt) {
			return pt, true;
		}
	}

	// There really doesn't seem to be any open positions.
	return Pt2I{0, 0}, false;
}


func (self *World) drawTerrain() {
	for pt := range PtIter(0, 0, mapWidth, mapHeight) {
		if self.GetLos(pt) == LosUnknown { continue; }
		tileset1[self.GetTerrain(pt)].Draw(pt.X, pt.Y);
	}
}

func (self *World) drawEntities() {
	for _, e := range self.entities {
		pos := e.GetPos();
		// TODO: Draw static (item) entities from map memory.
		if self.GetLos(pos) == LosSeen {
			e.Draw(pos.X, pos.Y);
		}
	}
}

func (self *World) getGuid(name string) (result Guid) {
	result = Guid(fmt.Sprintf("%v#%v", name, self.guidCounter));
	self.guidCounter++;
	return;
}