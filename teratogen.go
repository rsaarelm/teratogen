package main

import "container/vector"
import "exp/iterable"
import "fmt"
import "math"
import "rand"
import "time"

import "libtcod"
import . "fomalhaut"
import "sync"

// TODO: Librarize Point2
type IntPoint2 struct {
	X, Y int;
}

func (self *IntPoint2)GetIntPoint2() (x, y int) { return self.X, self.Y; }


// TODO: Librarize Rect
type IntRect struct {
	X, Y int;
	Width, Height int;
}

func (self *IntRect)RectArea() int { return self.Width * self.Height; }

func (self *IntRect)ContainsPoint(x, y int) bool {
	return x >= self.X && y >= self.Y &&
		x < self.X + self.Width && y < self.Y + self.Height;
}


// TODO: Librarize BspRoom.
type BspRoom struct {
	IntRect;
	ChildLeft, ChildRight *BspRoom;
}

func NewBspRoom(x, y int, w, h int) (result *BspRoom) {
	result = new(BspRoom);
	if w < 1 || h < 1 {
		Die("Making a BspRoom with zero dimension.");
	}
	result.X, result.Y, result.Width, result.Height = x, y, w, h;
	return;
}

func (self *BspRoom)IsLeaf() bool {
	return self.ChildLeft == nil && self.ChildRight == nil;
}

func (self *BspRoom)RoomAtPoint(x, y int) *BspRoom {
	if self.IsLeaf() {
		if self.IntRect.ContainsPoint(x, y) {
			return self;
		}
		return nil;
	} else {
		a := self.ChildLeft.RoomAtPoint(x, y);
		if a != nil { return a; }
		return self.ChildRight.RoomAtPoint(x, y);
	}
	panic("XXX: Issue 65");
}

func AddPointToConnectingWall(
	graph Graph, room1, room2 *BspRoom, x, y int) {
	arc, found := graph.GetArc(room1, room2);
	if !found {
		// These rooms aren't in the graph yet. Add a bidirectional
		// connection. Use a vector of points as the arc object. The
		// same object is aliased in both arc directions.
		arc = vector.New(0);
		graph.AddArc(room1, room2, arc);
		graph.AddArc(room2, room1, arc);
	}
	// Check for duplicate points
	ptVec := arc.(*vector.Vector);

	// Look for duplicates.
	for pt := range ptVec.Iter() {
		pt := pt.(*IntPoint2);
		// If one is found, return.
		if pt.X == x && pt.Y == y {
			return;
		}
	}
	// No duplicates, add the point to vector.
	ptVec.Push(&IntPoint2{x, y});
}

func (self *BspRoom)FindConnectingWalls(graph Graph) {
	for y := self.Y; y <= self.Y + self.Height; y++ {
		for x := self.X; x <= self.X + self.Width; x++ {
			// If the center point is a wall...
			if self.RoomAtPoint(x, y) == nil {
				// .. try to find two opposing room points and
				// two opposing wall points, which means it's
				// a wall between two rooms that could be
				// turned into a doorway.
				n := self.RoomAtPoint(x, y - 1);
				e := self.RoomAtPoint(x + 1, y);
				w := self.RoomAtPoint(x - 1, y);
				s := self.RoomAtPoint(x, y + 1);
				var room1, room2 *BspRoom;
				if n != nil && s != nil && n != s &&
					w == nil && e == nil {
					room1, room2 = n, s;
				}
				if e != nil && w != nil && e != w &&
					n == nil && s == nil {
					room1, room2 = e, w;
				}

				if room1 != nil && room2 != nil {
					AddPointToConnectingWall(
						graph, room1, room2, x, y);
				}
			}
		}
	}
}

func (self *BspRoom)VerticalSplit(pos int) {
	if !self.IsLeaf() {
		Die("Splitting a non-leaf BspRoom.");
	}
	if pos < 1 || pos > self.Height - 2 {
		Die("BspRoom split pos too close to wall.");
	}
	self.ChildLeft = NewBspRoom(
		self.X, self.Y, self.Width, pos);
	self.ChildRight = NewBspRoom(
		self.X, self.Y + pos + 1,
		self.Width, self.Height - pos - 1);
}

func (self *BspRoom)HorizontalSplit(pos int) {
	if !self.IsLeaf() {
		Die("Splitting a non-leaf BspRoom.");
	}
	if pos < 1 || pos > self.Width - 2 {
		Die("BspRoom split pos too close to wall.");
	}
	self.ChildLeft = NewBspRoom(
		self.X, self.Y, pos, self.Height);
	self.ChildRight = NewBspRoom(
		self.X + pos + 1, self.Y,
		self.Width - pos - 1, self.Height);
}

// Probability weight for vertical split, can't split below height 3.
func (self *BspRoom)VerticalSplitWeight() int { return IntMax(0, self.Height - 2); }

// Probability weight for horizontal split, can't split below width 3.
func (self *BspRoom)HorizontalSplitWeight() int { return IntMax(0, self.Width - 2); }

func MaybeSplitRoom(room *BspRoom) {
	const medianArea = 60.0;
	// Asymptotically approach 1 as room size grows. When size is
	// medianArea, chance to split is 50 %.
	splitProb := math.Atan(float64(room.RectArea()) / medianArea) / (0.5 * math.Pi);

	if WithProb(splitProb) {
		vw := room.VerticalSplitWeight();
		hw := room.HorizontalSplitWeight();
		if vw == 0 && hw == 0 {
			// Too small to split either way.
			return;
		}
		isVert := rand.Intn(vw + hw) < vw;
		if isVert {
			room.VerticalSplit(rand.Intn(room.Height - 3) + 1);
		} else {
			room.HorizontalSplit(rand.Intn(room.Width - 3) + 1);
		}

		// XXX: Could split these into goroutines, but then we'd need
		// to set up channels to signal when they are finished.
		MaybeSplitRoom(room.ChildLeft);
		MaybeSplitRoom(room.ChildRight);
	}
}

func MakeBspMap(x, y, w, h int) (result *BspRoom) {
	result = NewBspRoom(x, y, w, h);
	MaybeSplitRoom(result);
	return;
}

func wallsToMakeDoorsIn(wallGraph Graph) (result *vector.Vector) {
	const extraDoorProb = 0.2;

	result = vector.New(0);
	rooms := wallGraph.Nodes();
	connectedRooms := NewMapSet();
	edgeRooms := NewMapSet();

	// The room list comes from a map, the order should be reasonably
	// random so we don't need a specific rng op here.
	startRoom := rooms[0];
	connectedRooms.Add(startRoom);

	nextNodes, _ := wallGraph.Neighbors(startRoom);

	for _, e := range nextNodes {
		edgeRooms.Add(e);
	}

	for edgeRooms.Len() > 0 {
		// Pick a room connected by an edge to the current set of
		// connected rooms.
		nextRoom := RandomFromIterable(edgeRooms);

		// Since we've been lazy with the sets and haven't recorded
		// the walls by which the edge rooms are touching the set of
		// connected rooms, we'll just iterate through every wall of
		// the chosen edge room and punch the door through the first
		// to show up.

		rooms, walls := wallGraph.Neighbors(nextRoom);
		doorsMade := 0;
		for i := 0; i < len(rooms); i++ {
			if connectedRooms.Contains(rooms[i]) {
				if doorsMade == 0 || WithProb(extraDoorProb) {
					// Always make at least one door. With
					// some prob make doors to other
					// connected rooms to make the map
					// more interesting.
					result.Push(walls[i]);
					doorsMade++;
				}
			} else {
				edgeRooms.Add(rooms[i]);
			}
		}
		edgeRooms.Remove(nextRoom);
		connectedRooms.Add(nextRoom);
	}

	return;
}

func DoorLocations(wallGraph Graph) (result *vector.Vector) {
	result = vector.New(0);

	for wall := range wallsToMakeDoorsIn(wallGraph).Iter() {
		result.Push(RandomFromIterable(wall.(iterable.Iterable)));
	}

	return;
}

const tickerWidth = 80;

func updateTicker(str string, lineLength int) string {
	return PadString(EatPrefix(str, 1), lineLength);
}

type World struct {
	PlayerX, PlayerY int;
	Lock *sync.RWMutex;
}

func NewWorld() (result *World) {
	result = new(World);
	result.PlayerX = 40;
	result.PlayerY = 20;
	result.Lock = new(sync.RWMutex);
	return;
}

// TODO: Event system for changing world, event handler does lock/unlock, all
// changes in events. "Transactional database".

func (self *World) MovePlayer(dx, dy int) {
	self.Lock.Lock();
	defer self.Lock.Unlock();

	self.PlayerX += dx;
	self.PlayerY += dy;
}

func main() {
	fmt.Print("Welcome to Teratogen.\n");
	running := true;
	getch := make(chan byte);

	rand.Seed(time.Nanoseconds());

	libtcod.Init(80, 50, "Teratogen");
	libtcod.SetForeColor(libtcod.MakeColor(255, 255, 0));
	libtcod.PutChar(0, 0, 64, libtcod.BkgndNone);
	libtcod.PrintLeft(0, 2, libtcod.BkgndNone, "Hello, world!");
	libtcod.SetForeColor(libtcod.MakeColor(255, 0, 0));
	libtcod.PutChar(0, 0, 65, libtcod.BkgndNone);
	libtcod.Flush();
	world := NewWorld();

	area := MakeBspMap(1, 1, 78, 38);

	graph := NewSparseMatrixGraph();
	area.FindConnectingWalls(graph);
	doors := DoorLocations(graph);

	tickerLine := "                                                                                Teratogen online. ";

	go func() {
		for {
			const lettersAtTime = 1;
			const letterDelayNs = 1e9 * 0.20;
			// XXX: lettesDelayNs doesn't evaluate to an exact
			// integer due to rounding errors, and casting inexact
			// floats to integers is a compile-time error, so we
			// need an extra Floor operation here.
			time.Sleep(int64(math.Floor(letterDelayNs) * lettersAtTime));
			for x := 0; x <= lettersAtTime; x++ {
				tickerLine = updateTicker(tickerLine, tickerWidth);
			}
		}
	}();

	// Game logic
	go func() {
		for {
			key := <-getch;
			switch key {
			case 'q':
				running = false;
				// Colemak direction pad.
			case 'n':
				world.MovePlayer(-1, 0);
			case ',':
				world.MovePlayer(0, 1);
			case 'i':
				world.MovePlayer(1, 0);
			case 'u':
				world.MovePlayer(0, -1);
			case 'p':
				tickerLine += "Some text for the buffer... ";
			}
		}
	}();

	libtcod.SetForeColor(libtcod.MakeColor(0, 255, 0));
	for running {
		libtcod.Clear();

		for y := 0; y < 40; y++ {
			for x := 0; x < 80; x++ {
				if area.RoomAtPoint(x, y) != nil {
					libtcod.SetForeColor(libtcod.MakeColor(96, 96, 96));
					libtcod.PutChar(x, y + 1, '.', libtcod.BkgndNone);
				} else {
					libtcod.SetForeColor(libtcod.MakeColor(192, 192, 0));
					libtcod.PutChar(x, y + 1, '#', libtcod.BkgndNone);
				}
			}
		}

		libtcod.SetForeColor(libtcod.MakeColor(0, 255, 255));
		for pt := range doors.Iter() {
			pt := pt.(*IntPoint2);
			libtcod.PutChar(pt.X, pt.Y + 1, '+', libtcod.BkgndNone);
		}

		libtcod.SetForeColor(libtcod.MakeColor(0, 255, 0));

		libtcod.PutChar(world.PlayerX, world.PlayerY + 1, '@', libtcod.BkgndNone);

		libtcod.SetForeColor(libtcod.MakeColor(192, 192, 192));
		libtcod.PrintLeft(0, 0, libtcod.BkgndNone, tickerLine);

		libtcod.Flush();

		key := libtcod.CheckForKeypress();
		if key != 0 {
			getch <- byte(key);
		}
	}
}
