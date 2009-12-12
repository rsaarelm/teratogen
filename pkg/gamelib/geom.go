package gamelib

import (
	"math"
)

type Vec2I struct {
	X	int
	Y	int
}

type Pt2I Vec2I

func (lhs Vec2I) Equals(rhs Vec2I) bool	{ return lhs.X == rhs.X && lhs.Y == rhs.Y }

func (lhs Pt2I) Equals(rhs Pt2I) bool	{ return lhs.X == rhs.X && lhs.Y == rhs.Y }

func (lhs Vec2I) Plus(rhs Vec2I) (result Vec2I) {
	return Vec2I{lhs.X + rhs.X, lhs.Y + rhs.Y}
}

func (lhs Pt2I) Plus(rhs Vec2I) (result Pt2I)	{ return Pt2I{lhs.X + rhs.X, lhs.Y + rhs.Y} }

func (lhs Vec2I) Minus(rhs Vec2I) (result Vec2I) {
	return Vec2I{lhs.X - rhs.X, lhs.Y - rhs.Y}
}

func (lhs Pt2I) Minus(rhs Pt2I) (result Vec2I) {
	return Vec2I{lhs.X - rhs.X, lhs.Y - rhs.Y}
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

func (lhs Vec2I) Dot(rhs Vec2I) int	{ return lhs.X*rhs.X + lhs.Y*rhs.Y }

func (self Vec2I) Abs() float64	{ return math.Sqrt(float64(self.Dot(self))) }

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

type RectI struct {
	Pos	Pt2I
	Dim	Vec2I
}

func (self RectI) Contains(pos Pt2I) bool {
	return pos.X >= self.Pos.X && pos.Y >= self.Pos.Y &&
		pos.X < self.Pos.X+self.Dim.X &&
		pos.Y < self.Pos.Y+self.Dim.Y
}

func (self RectI) RectArea() int	{ return self.Dim.X * self.Dim.Y }

func (self RectI) Iter() <-chan Pt2I {
	return PtIter(self.Pos.X, self.Pos.Y, self.Dim.X, self.Dim.Y)
}

func (self RectI) X() int { return self.Pos.X }

func (self RectI) Y() int { return self.Pos.Y }

func (self RectI) Width() int { return self.Dim.X }

func (self RectI) Height() int { return self.Dim.Y }
