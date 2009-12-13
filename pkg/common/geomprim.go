package common

import (
	"math"
)

// Vector type in common package because to reduce typing with its frequent
// use.
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
