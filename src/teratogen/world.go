package teratogen

//import "fmt"
import "rand"
import "sync"

import . "fomalhaut"

const mapWidth = 80
const mapHeight = 40

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


type Entity interface {
	// TODO: Entity-common stuff.
	IsObstacle() bool;
	GetPos() (int, int);
}

type Guid string


type Creature struct {
	Icon;
	Name string;
	X, Y int;
}

type World struct {
	PlayerX, PlayerY int;
	Lock *sync.RWMutex;
	entities map[Guid] Entity;
	terrain Field2;
}

func NewWorld() (result *World) {
	result = new(World);
	result.PlayerX = 40;
	result.PlayerY = 20;
	result.entities = make(map[Guid] Entity);
	result.terrain = NewMapField2();
	result.Lock = new(sync.RWMutex);
	return;
}

// TODO: Event system for changing world, event handler does lock/unlock, all
// changes in events. "Transactional database".

func (self *World) MovePlayer(dx, dy int) {
	self.Lock.Lock();
	defer self.Lock.Unlock();

	newX := self.PlayerX + dx;
	newY := self.PlayerY + dy;

	if self.IsOpen(newX, newY) {
		self.PlayerX, self.PlayerY = newX, newY;
	}
}

func (self *World)InitLevel(num int, player Entity) {
	self.terrain = NewMapField2();
	self.entities = make(map[Guid] Entity);

	self.makeBSPMap();

	x, y, ok := self.GetMatchingPos(
		func(x, y int) bool { return self.IsSpawnPos(x, y); });
	if !ok {
		Die("Couldn't find open position.");
	}
	self.PlayerX = x;
	self.PlayerY = y;
}

func (self *World)makeBSPMap() {
	area := MakeBspMap(1, 1, mapWidth - 2, mapHeight - 2);
	graph := NewSparseMatrixGraph();
	area.FindConnectingWalls(graph);
	doors := DoorLocations(graph);

	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			if area.RoomAtPoint(x, y) != nil {
				self.terrain.Set(x, y, TerrainFloor);
			} else {
				self.terrain.Set(x, y, TerrainWall);
			}
		}
	}

	for pt := range doors.Iter() {
		pt := pt.(*IntPoint2);
		// TODO: Door terrain
		self.terrain.Set(pt.X, pt.Y, TerrainDoor);
	}

}

func (self *World)GetTerrain(x, y int) TerrainType {
	if val, ok := self.terrain.Get(x, y); ok {
		// XXX: Can't cast it straight to TerrainType for some reason.
		return TerrainType(val.(int));
	}
	return TerrainIndeterminate;
}


func (self *World)EntitiesAt(x, y int) <-chan Entity {
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


func (self *World)IsOpen(x, y int) bool {
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

func (self *World)IsSpawnPos(x, y int) bool {
	if !self.IsOpen(x, y) { return false; }
	if self.GetTerrain(x, y) == TerrainDoor { return false; }
	return true;
}

func (self *World)GetMatchingPos(f func (x, y int) bool) (x, y int, found bool) {
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

func (self *World)Draw() {
	self.DrawTerrain();
	self.DrawEntities();
}

func (self *World)DrawTerrain() {
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			tileset1[self.GetTerrain(x, y)].Draw(x, y);
		}
	}
}

func (self *World)DrawEntities() {
	// TODO: Proper entities instead of this hack.
	DrawCharRGB(self.PlayerX, self.PlayerY, '@', RGB{0, 255, 0});
}