package gamelib

import "math"

func LineOfSight(
	isBlocked func(Vec2I) bool,
	outsideRadius func(Vec2I) bool) <-chan Vec2I {
	c := make(chan Vec2I);

	go func() {
		c <- Vec2I{0, 0};

		for octant := 0; octant < 8; octant++ {
			processOctant(
				c, isBlocked, outsideRadius, octant,
				0.0, 1.0, 1);
		}
		close(c);
	}();

	return c;
}

func processOctant(
	c chan<- Vec2I,
	isBlocked func(Vec2I) bool,
	outsideRadius func(Vec2I) bool,
	octant int,
	startSlope float64,
	endSlope float64,
	u int) {
	if outsideRadius(Vec2I{u, 0}) {
		return;
	}

	if endSlope <= startSlope {
		return;
	}

	traversingObstacle := true;

	for v, ev := int(Round(float64(u) * startSlope)), int(math.Ceil(float64(u) * endSlope));
	    v <= ev; v++ {
		var pos Vec2I;
		switch octant {
		case 0: pos = Vec2I{v, -u};
		case 1: pos = Vec2I{u, -v};
		case 2: pos = Vec2I{u, v};
		case 3: pos = Vec2I{v, u};
		case 4: pos = Vec2I{-v, u};
		case 5: pos = Vec2I{-u, v};
		case 6: pos = Vec2I{-u, -v};
		case 7: pos = Vec2I{-v, -u};
		default: Die("Bad octant %v", octant);
		}

		if isBlocked(pos) {
			if !traversingObstacle {
				// Hit an obstacle.
				processOctant(
					c,
					isBlocked,
					outsideRadius,
					octant,
					startSlope,
					(float64(v) - 0.5) / (float64(u) + 0.5),
					u + 1);
				traversingObstacle = true;
			}
		} else {
			if traversingObstacle {
				// Risen above an obstacle.
				traversingObstacle = false;
				if (v > 0) {
					startSlope = Float64Max(
						startSlope,
						(float64(v) - 0.5) / (float64(u) - 0.5));
				}
			}
		}

		if startSlope < endSlope && !outsideRadius(pos) {
			c <- pos;
		}
        }

	if !traversingObstacle {
		processOctant(
			c,
			isBlocked,
			outsideRadius,
			octant,
			startSlope,
			endSlope,
			u + 1);
	}
}

// Determine the 1/16th sector of a circle a point in the XY plane points
// towards. Sector 0 is clockwise from the y-axis, and subsequent sectors are
// clockwise from there. The origin point is handled in the same way as in
// math.Atan2.
func Hexadecant(x, y float64) int {
	const hexadecantWidth = math.Pi / 8.0;
	radian := math.Atan2(x, -y);
	if radian < 0 { radian += 2.0 * math.Pi }
	return int(math.Floor(radian / hexadecantWidth));
}

func Vec2IToDir8(vec Vec2I) int {
	return ((Hexadecant(float64(vec.X), float64(vec.Y)) + 1) % 16) / 2;
}

func Vec2IToDir4(vec Vec2I) int {
	return ((Hexadecant(float64(vec.X), float64(vec.Y)) + 2) % 16) / 4;
}

func Dir8ToVec(dir int) Vec2I {
	switch dir {
	case 0: return Vec2I{0, -1};
	case 1: return Vec2I{1, -1};
	case 2: return Vec2I{1, 0};
	case 3: return Vec2I{1, 1};
	case 4: return Vec2I{0, 1};
	case 5: return Vec2I{-1, 1};
	case 6: return Vec2I{-1, 0};
	case 7: return Vec2I{-1, -1};
	}
	panic("Invalid dir");
}

func PosAdjacent(p1, p2 Pt2I) bool {
	diff := p1.Minus(p2);
	x, y := Iabs(diff.X), Iabs(diff.Y);
	return x < 2 && y < 2 && x + y > 0;
}