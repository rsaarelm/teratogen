package fomalhaut

import "math"

type Vec2I struct {
	X int;
	Y int;
}

func (lhs Vec2I) Equals(rhs Vec2I) bool {
	return lhs.X == rhs.X && lhs.Y == rhs.Y;
}

func (lhs Vec2I) Plus(rhs Vec2I) (result Vec2I) {
	return Vec2I{lhs.X + rhs.X, lhs.Y + rhs.Y};
}

func (lhs Vec2I) Minus(rhs Vec2I) (result Vec2I) {
	return Vec2I{lhs.X - rhs.X, lhs.Y - rhs.Y};
}

func (lhs Vec2I) Dot(rhs Vec2I) int {
	return lhs.X * rhs.X + lhs.Y * rhs.Y;
}

func (self Vec2I) Abs() float64 {
	return math.Sqrt(float64(self.Dot(self)));
}

// Iterate points where 0 <= x < self.X and 0 <= y < self.Y.
func (self Vec2I) Iter() <-chan Vec2I {
	c := make(chan Vec2I);
	go func() {
		for y := 0; y < self.Y; y++ {
			for x:= 0; x < self.X; x++ {
				c <- Vec2I{x, y};
			}
		}
		close(c);
	}();
	return c;
}