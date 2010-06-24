package geom

import (
	"exp/iterable"
	"math"
)

type Vec2I struct {
	X int
	Y int
}

type Pt2I Vec2I

var ZeroVec2I = Vec2I{0, 0}

var Origin = Pt2I{0, 0}

func (lhs Vec2I) Equals(rhs Vec2I) bool { return lhs.X == rhs.X && lhs.Y == rhs.Y }

func (lhs Pt2I) Equals(rhs Pt2I) bool { return lhs.X == rhs.X && lhs.Y == rhs.Y }

func (lhs Vec2I) Plus(rhs Vec2I) (result Vec2I) {
	return Vec2I{lhs.X + rhs.X, lhs.Y + rhs.Y}
}

func (self Vec2I) Neg() Vec2I { return Vec2I{-self.X, -self.Y} }

func (lhs Pt2I) Plus(rhs Vec2I) (result Pt2I) { return Pt2I{lhs.X + rhs.X, lhs.Y + rhs.Y} }

func (lhs Vec2I) Minus(rhs Vec2I) (result Vec2I) {
	return Vec2I{lhs.X - rhs.X, lhs.Y - rhs.Y}
}

func (lhs Pt2I) Minus(rhs Pt2I) (result Vec2I) {
	return Vec2I{lhs.X - rhs.X, lhs.Y - rhs.Y}
}

func (lhs Vec2I) ElemMult(rhs Vec2I) (result Vec2I) {
	return Vec2I{lhs.X * rhs.X, lhs.Y * rhs.Y}
}

func (lhs Vec2I) Scale(rhs int) (result Vec2I) {
	return Vec2I{lhs.X * rhs, lhs.Y * rhs}
}

func (lhs Pt2I) ElemMult(rhs Vec2I) (result Pt2I) {
	return Pt2I{lhs.X * rhs.X, lhs.Y * rhs.Y}
}

func (self *Pt2I) Add(rhs Vec2I) {
	self.X += rhs.X
	self.Y += rhs.Y
}

func (self *Pt2I) Subtract(rhs Vec2I) {
	self.X -= rhs.X
	self.Y -= rhs.Y
}

func (self *Vec2I) Add(rhs Vec2I) {
	self.X += rhs.X
	self.Y += rhs.Y
}

func (self *Vec2I) Subtract(rhs Vec2I) {
	self.X -= rhs.X
	self.Y -= rhs.Y
}

func (lhs Vec2I) Dot(rhs Vec2I) int { return lhs.X*rhs.X + lhs.Y*rhs.Y }

func (self Vec2I) Abs() float64 { return math.Sqrt(float64(self.Dot(self))) }

// XXX: Should use iterable.Iterable?

// Iterate points where x0 <= x < x0 + width and y0 <= y < y0 + height.
func PtIter(x0, y0, width, height int) <-chan Pt2I {
	c := make(chan Pt2I)
	go func() {
		for y := y0; y < y0+height; y++ {
			for x := x0; x < x0+width; x++ {
				c <- Pt2I{x, y}
			}
		}
		close(c)
	}()
	return c
}

// Iterate through points at the edge of a rectangle.
func EdgeIter(x0, y0, width, height int) iterable.Iterable {
	x1 := x0 + width - 1
	y1 := y0 + height - 1
	return iterable.Func(func(c chan<- interface{}) {
		for x := x0; x < x0+width; x++ {
			c <- Pt2I{x, y0}
		}
		for y := y0 + 1; y < y0+height; y++ {
			c <- Pt2I{x1, y}
		}
		for x := x1 - 1; x >= x0; x-- {
			c <- Pt2I{x, y1}
		}
		for y := y1 - 1; y > y0; y-- {
			c <- Pt2I{x0, y}
		}
		close(c)
	})
}

func IsCorner(x0, y0, width, height int, pos Pt2I) bool {
	return (pos.X == x0 || pos.X == x0+width-1) &&
		(pos.Y == y0 || pos.Y == y0+height-1)
}

func IsAtEdge(x0, y0, width, height int, pos Pt2I) bool {
	return (pos.X == x0 || pos.X == x0+width-1) ||
		(pos.Y == y0 || pos.Y == y0+height-1)
}

type RectI struct {
	Pos Pt2I
	Dim Vec2I
}

func (self RectI) Contains(pos Pt2I) bool {
	return pos.X >= self.Pos.X && pos.Y >= self.Pos.Y &&
		pos.X < self.Pos.X+self.Dim.X &&
		pos.Y < self.Pos.Y+self.Dim.Y
}

func (self RectI) RectArea() int { return self.Dim.X * self.Dim.Y }

// XXX: Should use iterable.Iterable?

func (self RectI) Iter() <-chan Pt2I {
	return PtIter(self.Pos.X, self.Pos.Y, self.Dim.X, self.Dim.Y)
}

func (self RectI) X() int { return self.Pos.X }

func (self RectI) Y() int { return self.Pos.Y }

func (self RectI) Width() int { return self.Dim.X }

func (self RectI) Height() int { return self.Dim.Y }

// CenterRect gives the X, Y offset for rectangle inner such that the centers
// of rectangles inner and outer match when outer has offset (0, 0).
func CenterRects(innerW, innerH int, outerW, outerH int) (innerX, innerY int) {
	return (outerW - innerW) / 2, (outerH - innerH) / 2
}
