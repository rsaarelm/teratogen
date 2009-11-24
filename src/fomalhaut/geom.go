package fomalhaut

import "math"

type Vec2I [2]int;

func (lhs Vec2I) Equals(rhs Vec2I) bool {
	return lhs[0] == rhs[0] && lhs[1] == rhs[1];
}

func (lhs Vec2I) Plus(rhs Vec2I) (result Vec2I) {
	return Vec2I{lhs[0] + rhs[0], lhs[1] + rhs[1]};
}

func (lhs Vec2I) Minus(rhs Vec2I) (result Vec2I) {
	return Vec2I{lhs[0] - rhs[0], lhs[1] - rhs[1]};
}

func (lhs Vec2I) Dot(rhs Vec2I) int {
	return lhs[0] * rhs[0] + lhs[1] * rhs[1];
}

func (self Vec2I) Abs() float64 {
	return math.Sqrt(float64(self.Dot(self)));
}

// Iterate points where 0 <= x < self[0] and 0 <= y < self[1].
func (self Vec2I) Iter() <-chan Vec2I {
	c := make(chan Vec2I);
	go func() {
		for y := 0; y < self[1]; y++ {
			for x:= 0; x < self[0]; x++ {
				c <- Vec2I{x, y};
			}
		}
		close(c);
	}();
	return c;
}