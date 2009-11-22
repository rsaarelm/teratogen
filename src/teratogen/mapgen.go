package teratogen

import "container/vector"
import "exp/iterable"
import "math"
import "rand"

import . "fomalhaut"

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
	// The higher this is, the more the splitter will tend to pick a
	// direction that brings subroom shapes closer to a square.
	const aspectNormalizingExponent = 2.0;

	// Split probability approaches 1 asymptotically as room size grows.
	// When size is medianArea, chance to split is 50 %.
	const medianArea = 60.0;
	splitProb := math.Atan(float64(room.RectArea()) / medianArea) / (0.5 * math.Pi);

	if WithProb(splitProb) {
		vw := int(math.Pow(
			float64(room.VerticalSplitWeight()),
			aspectNormalizingExponent));
		hw := int(math.Pow(
			float64(room.HorizontalSplitWeight()),
			aspectNormalizingExponent));
		if vw < 1 && hw < 1 {
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
	rooms := iterable.Data(wallGraph);

	if len(rooms) == 0 {
		return;
	}

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