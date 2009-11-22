package teratogen

type IntPoint2 struct {
	X, Y int;
}

func (self *IntPoint2)GetIntPoint2() (x, y int) { return self.X, self.Y; }


type IntRect struct {
	X, Y int;
	Width, Height int;
}

func (self *IntRect)RectArea() int { return self.Width * self.Height; }

func (self *IntRect)ContainsPoint(x, y int) bool {
	return x >= self.X && y >= self.Y &&
		x < self.X + self.Width && y < self.Y + self.Height;
}
