package geom

import (
	. "hyades/common"
)

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
