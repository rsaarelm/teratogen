package main

import (
	"container/vector"
	"exp/iterable"
	. "hyades/gamelib"
	"hyades/geom"
	"hyades/num"
	"math"
	"rand"
)


const minRoomDim = 2

type BspRoom struct {
	geom.RectI
	ChildLeft, ChildRight	*BspRoom
}

func NewBspRoom(x, y int, w, h int) (result *BspRoom) {
	result = new(BspRoom)
	Assert(w > 0 && h > 0, "Making a BspRoom with zero dimension.")
	result.Pos = geom.Pt2I{x, y}
	result.Dim = geom.Vec2I{w, h}
	return
}

func (self *BspRoom) IsLeaf() bool	{ return self.ChildLeft == nil && self.ChildRight == nil }

func (self *BspRoom) RoomAtPoint(x, y int) *BspRoom {
	if self.IsLeaf() {
		if self.Contains(geom.Pt2I{x, y}) {
			return self
		}
		return nil
	} else {
		a := self.ChildLeft.RoomAtPoint(x, y)
		if a != nil {
			return a
		}
		return self.ChildRight.RoomAtPoint(x, y)
	}
	panic("XXX: Issue 65")
}

func AddPointToConnectingWall(graph Graph, room1, room2 *BspRoom, x, y int) {
	arc, found := graph.GetArc(room1, room2)
	if !found {
		// These rooms aren't in the graph yet. Add a bidirectional
		// connection. Use a vector of points as the arc object. The
		// same object is aliased in both arc directions.
		arc = new(vector.Vector)
		graph.AddArc(room1, room2, arc)
		graph.AddArc(room2, room1, arc)
	}
	// Check for duplicate points
	ptVec := arc.(*vector.Vector)

	// Look for duplicates.
	for pt := range ptVec.Iter() {
		pt := pt.(geom.Pt2I)
		// If one is found, return.
		if pt.X == x && pt.Y == y {
			return
		}
	}
	// No duplicates, add the point to vector.
	ptVec.Push(geom.Pt2I{x, y})
}

func (self *BspRoom) FindConnectingWalls(graph Graph) {
	for pt := range self.Iter() {
		// If the center point is a wall...
		if self.RoomAtPoint(pt.X, pt.Y) == nil {
			// .. try to find two opposing room points and
			// two opposing wall points, which means it's
			// a wall between two rooms that could be
			// turned into a doorway.
			n := self.RoomAtPoint(pt.X, pt.Y-1)
			e := self.RoomAtPoint(pt.X+1, pt.Y)
			w := self.RoomAtPoint(pt.X-1, pt.Y)
			s := self.RoomAtPoint(pt.X, pt.Y+1)
			var room1, room2 *BspRoom
			if n != nil && s != nil && n != s &&
				w == nil && e == nil {
				room1, room2 = n, s
			}
			if e != nil && w != nil && e != w &&
				n == nil && s == nil {
				room1, room2 = e, w
			}

			if room1 != nil && room2 != nil {
				AddPointToConnectingWall(
					graph, room1, room2, pt.X, pt.Y)
			}
		}
	}
}

func (self *BspRoom) VerticalSplit(pos int) {
	Assert(self.IsLeaf(), "Splitting a non-leaf BspRoom.")
	Assert(pos >= minRoomDim && pos < self.Dim.Y-minRoomDim,
		"BspRoom split pos too close to wall.")
	self.ChildLeft = NewBspRoom(
		self.Pos.X, self.Pos.Y, self.Dim.X, pos)
	self.ChildRight = NewBspRoom(
		self.Pos.X, self.Pos.Y+pos+1,
		self.Dim.X, self.Dim.Y-pos-1)
}

func (self *BspRoom) HorizontalSplit(pos int) {
	Assert(self.IsLeaf(), "Splitting a non-leaf BspRoom.")
	Assert(pos >= minRoomDim && pos < self.Dim.X-minRoomDim,
		"BspRoom split pos too close to wall.")
	self.ChildLeft = NewBspRoom(
		self.Pos.X, self.Pos.Y, pos, self.Dim.Y)
	self.ChildRight = NewBspRoom(
		self.Pos.X+pos+1, self.Pos.Y,
		self.Dim.X-pos-1, self.Dim.Y)
}

// Probability weight for vertical split, can't split below height minRoomDim * 2 + 1.
func (self *BspRoom) VerticalSplitWeight() int {
	return num.IntMax(0, self.Dim.Y-minRoomDim*2)
}

// Probability weight for horizontal split, can't split below width minRoomDim * 2 + 1.
func (self *BspRoom) HorizontalSplitWeight() int {
	return num.IntMax(0, self.Dim.X-minRoomDim*2)
}

func MaybeSplitRoom(room *BspRoom) {
	// The higher this is, the more the splitter will tend to pick a
	// direction that brings subroom shapes closer to a square.
	const aspectNormalizingExponent = 2.0

	// Split probability approaches 1 asymptotically as room size grows.
	// When size is medianArea, chance to split is 50 %.
	const medianArea = 60.0
	splitProb := math.Atan(float64(room.RectArea())/medianArea) / (0.5 * math.Pi)

	if num.WithProb(splitProb) {
		vw := int(math.Pow(
			float64(room.VerticalSplitWeight()),
			aspectNormalizingExponent))
		hw := int(math.Pow(
			float64(room.HorizontalSplitWeight()),
			aspectNormalizingExponent))
		if vw < 1 && hw < 1 {
			// Too small to split either way.
			return
		}
		isVert := rand.Intn(vw+hw) < vw
		if isVert {
			// Do two random calls to concentrate distribution
			// around the middle. The (span + 1) bit in the second
			// one is a trick to get the whole range even when
			// span is odd and gets truncated by integer division.
			span := room.Dim.Y - (2*minRoomDim + 1)
			splitPos := rand.Intn(span/2) + rand.Intn((span+1)/2) + minRoomDim
			room.VerticalSplit(splitPos)
		} else {
			span := room.Dim.X - (2*minRoomDim + 1)
			splitPos := rand.Intn(span/2) + rand.Intn((span+1)/2) + minRoomDim
			room.HorizontalSplit(splitPos)
		}

		// XXX: Could split these into goroutines, but then we'd need
		// to set up channels to signal when they are finished.
		MaybeSplitRoom(room.ChildLeft)
		MaybeSplitRoom(room.ChildRight)
	}
}

func MakeBspMap(x, y, w, h int) (result *BspRoom) {
	result = NewBspRoom(x, y, w, h)
	MaybeSplitRoom(result)
	return
}

func wallsToMakeDoorsIn(wallGraph Graph) (result *vector.Vector) {
	const extraDoorProb = 0.2

	result = new(vector.Vector)
	rooms := iterable.Data(wallGraph)

	if len(rooms) == 0 {
		return
	}

	connectedRooms := NewMapSet()
	edgeRooms := NewMapSet()

	// The room list comes from a map, the order should be reasonably
	// random so we don't need a specific rng op here.
	startRoom := rooms[0]
	connectedRooms.Add(startRoom)

	nextNodes, _ := wallGraph.Neighbors(startRoom)

	for _, e := range nextNodes {
		edgeRooms.Add(e)
	}

	for edgeRooms.Len() > 0 {
		// Pick a room connected by an edge to the current set of
		// connected rooms.
		nextRoom := num.RandomFromIterable(edgeRooms)

		// Since we've been lazy with the sets and haven't recorded
		// the walls by which the edge rooms are touching the set of
		// connected rooms, we'll just iterate through every wall of
		// the chosen edge room and punch the door through the first
		// to show up.

		rooms, walls := wallGraph.Neighbors(nextRoom)
		doorsMade := 0
		for i := 0; i < len(rooms); i++ {
			if connectedRooms.Contains(rooms[i]) {
				if doorsMade == 0 || num.WithProb(extraDoorProb) {
					// Always make at least one door. With
					// some prob make doors to other
					// connected rooms to make the map
					// more interesting.
					result.Push(walls[i])
					doorsMade++
				}
			} else {
				edgeRooms.Add(rooms[i])
			}
		}
		edgeRooms.Remove(nextRoom)
		connectedRooms.Add(nextRoom)
	}

	return
}

func DoorLocations(wallGraph Graph) (result *vector.Vector) {
	result = new(vector.Vector)

	for wall := range wallsToMakeDoorsIn(wallGraph).Iter() {
		result.Push(num.RandomFromIterable(wall.(iterable.Iterable)))
	}

	return
}

type CaveTile byte

const (
	CaveUnknown	= iota
	CaveFloor
	CaveWall
)

// Cave generator by Ray Dillinger, Message-Id: <48d8aa27$0$33580$742ec2ed@news.sonic.net>
// Adapted from the original C to Golang.
func MakeCaveMap(width, height int, floorPercent float) (result [][]CaveTile) {
	const iterationsPerCell = 500
	const recarveProb = 0.01
	maxFloorCount := int(floorPercent * float(width*height))

	result = make([][]CaveTile, width)
	for x := 0; x < width; x++ {
		result[x] = make([]CaveTile, height)
	}

	uncommittedCount := width*height - 1
	wallCount := 0
	floorCount := 1

	xmin, ymin := width/2-1, height/2-1
	xmax, ymax := xmin+2, ymin+2

	iterationLimit := 0

	// Clear a center starting point.
	result[width/2][height/2] = CaveFloor

	for iterationLimit < width*height*iterationsPerCell && floorCount < maxFloorCount {
		iterationLimit++
		x, y := xmin+rand.Intn(xmax-xmin+1), ymin+rand.Intn(ymax-ymin+1)
		if result[x][y] == CaveUnknown || num.WithProb(recarveProb) {
			if x == xmin && x > 1 {
				xmin--
			}
			if x == xmax && x < width-2 {
				xmax++
			}
			if y == ymin && y > 1 {
				ymin--
			}
			if y == ymax && y < height-2 {
				ymax++
			}

			adjFloors := 0
			if result[x-1][y] == CaveFloor {
				adjFloors++
			}
			if result[x+1][y] == CaveFloor {
				adjFloors++
			}
			if result[x][y-1] == CaveFloor {
				adjFloors++
			}
			if result[x][y+1] == CaveFloor {
				adjFloors++
			}

			adjWalls := 0
			if result[x-1][y] == CaveWall {
				adjWalls++
			}
			if result[x+1][y] == CaveWall {
				adjWalls++
			}
			if result[x][y-1] == CaveWall {
				adjWalls++
			}
			if result[x][y+1] == CaveWall {
				adjWalls++
			}

			if adjFloors > 0 {
				if uncommittedCount+floorCount > width*height/2 &&
					(adjWalls > adjFloors || wallCount*3 < floorCount*2) {
					if result[x][y] == CaveUnknown {
						uncommittedCount--
					}
					if result[x][y] == CaveFloor {
						floorCount--
					}
					if result[x][y] != CaveWall {
						wallCount++
					}
					result[x][y] = CaveWall
				} else {
					if result[x][y] == CaveUnknown {
						uncommittedCount--
					}
					if result[x][y] == CaveWall {
						wallCount--
					}
					if result[x][y] != CaveFloor {
						floorCount++
					}
					result[x][y] = CaveFloor
				}
			}
		}
	}

	for x := 1; x < width-1; x++ {
		for y := 1; y < height-1; y++ {
			adjFloors := 0
			if result[x-1][y] == CaveFloor {
				adjFloors++
			}
			if result[x+1][y] == CaveFloor {
				adjFloors++
			}
			if result[x][y-1] == CaveFloor {
				adjFloors++
			}
			if result[x][y+1] == CaveFloor {
				adjFloors++
			}

			if adjFloors > 0 && result[x][y] == CaveUnknown {
				result[x][y] = CaveWall
			}
			if adjFloors == 4 {
				result[x][y] = CaveFloor
			}
			if adjFloors == 0 {
				result[x][y] = CaveWall
			}
		}
	}

	return
}
