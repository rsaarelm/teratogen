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
	GetPos() (int, int);
	GetGuid() Guid;
	MoveAbs(x, y int);
	Move(dx, dy int);
}


type Creature struct {
	*Icon;
	guid Guid;
	Name string;
	x, y int;
}

func (self *Creature) IsObstacle() bool { return true; }

func (self *Creature) GetPos() (x, y int) { return self.x, self.y; }

func (self *Creature) GetGuid() Guid { return self.guid; }

func (self *Creature) MoveAbs(x, y int) { self.x, self.y = x, y; }

func (self *Creature) Move(dx, dy int) {
	x, y := self.GetPos();
	self.MoveAbs(x + dx, y + dy);
}


type World struct {
	playerId Guid;
	Lock *sync.RWMutex;
	entities map[Guid] Entity;
	terrain []TerrainType;
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
		result = &Creature{&Icon{'@', RGB{0xdd, 0xff, 0xff}}, guid, "protagonist", -1, -1};
	case EntityZombie:
		result = &Creature{&Icon{'z', RGB{0x80, 0xa0, 0x80}}, guid, "zombie", -1, -1};
	default:
		Die("Unknown entity type.");
	}
	self.entities[guid] = result;
	return;
}

func (self *World) SpawnAt(entityType EntityType, x, y int) (result Entity) {
	result = self.Spawn(entityType);
	result.MoveAbs(x, y);
	return;
}

func (self *World) SpawnRandomPos(entityType EntityType) (result Entity) {
	x, y := self.GetSpawnPos();
	return self.SpawnAt(entityType, x, y);
}

// TODO: Event system for changing world, event handler does lock/unlock, all
// changes in events. "Transactional database".

func (self *World) MovePlayer(dx, dy int) {
	self.Lock.Lock();
	defer self.Lock.Unlock();

	player := self.GetPlayer();
	x, y := player.GetPos();

	if self.IsOpen(x + dx, y + dy) {
		player.Move(dx, dy)
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

	x, y := self.GetSpawnPos();
	player.MoveAbs(x, y);

	for i := 0; i < 10 + num * 4; i++ {
		self.SpawnRandomPos(EntityZombie);
	}
}

func (self *World) initTerrain() {
	self.terrain = make([]TerrainType, numTerrainCells);
}

func (self *World) makeBSPMap() {
	area := MakeBspMap(1, 1, mapWidth - 2, mapHeight - 2);
	graph := NewSparseMatrixGraph();
	area.FindConnectingWalls(graph);
	doors := DoorLocations(graph);

	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			if area.RoomAtPoint(x, y) != nil {
				self.SetTerrain(x, y, TerrainFloor);
			} else {
				self.SetTerrain(x, y, TerrainWall);
			}
		}
	}

	for pt := range doors.Iter() {
		pt := pt.(*IntPoint2);
		// TODO: Door terrain
		self.SetTerrain(pt.X, pt.Y, TerrainDoor);
	}

}

func inTerrain(x, y int) bool {
	return x >= 0 && y >= 0 && x < mapWidth && y < mapHeight;
}

func (self *World) GetTerrain(x, y int) TerrainType {
	if inTerrain(x, y) {
		return self.terrain[x + y * mapWidth];
	}
	return TerrainIndeterminate;
}

func (self *World) SetTerrain(x, y int, t TerrainType) {
	if inTerrain(x, y) {
		self.terrain[x + y * mapWidth] = t
	}
}

func (self *World) EntitiesAt(x, y int) <-chan Entity {
	c := make(chan Entity);
	go func (x, y int, c chan<- Entity) {
		for _, ent := range self.entities {
			entX, entY := ent.GetPos();
			if entX == x && entY == y {
				c <- ent;
			}
		}
		close(c);
	}(x, y, c);
	return c;
}


func (self *World) IsOpen(x, y int) bool {
	if IsObstacleTerrain(self.GetTerrain(x, y)) {
		return false;
	}
	for e := range self.EntitiesAt(x, y) {
		if e.IsObstacle() {
			return false;
		}
	}

	return true;
}

func (self *World) GetSpawnPos() (x, y int) {
	x, y, ok := self.GetMatchingPos(
		func(x, y int) bool { return self.isSpawnPos(x, y); });
	if !ok {
		Die("Couldn't find open position.");
	}
	return;
}

func (self *World) isSpawnPos(x, y int) bool {
	if !self.IsOpen(x, y) { return false; }
	if self.GetTerrain(x, y) == TerrainDoor { return false; }
	return true;
}

func (self *World) GetMatchingPos(f func (x, y int) bool) (x, y int, found bool) {
	const tries = 1024;

	for i := 0; i < tries; i++ {
		x, y := rand.Intn(mapWidth), rand.Intn(mapHeight);
		if f(x, y) {
			return x, y, true;
		}
	}

	// RNG has failed us, let's do an exhaustive search..
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			if f(x, y) {
				return x, y, true;
			}
		}
	}

	// There really doesn't seem to be any open positions.
	return 0, 0, false;
}


func (self *World) drawTerrain() {
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			tileset1[self.GetTerrain(x, y)].Draw(x, y);
		}
	}
}

func (self *World) drawEntities() {
	for _, e := range self.entities {
		x, y := e.GetPos();
		e.Draw(x, y);
	}
}

func (self *World) getGuid(name string) (result Guid) {
	result = Guid(fmt.Sprintf("%v#%v", name, self.guidCounter));
	self.guidCounter++;
	return;
}