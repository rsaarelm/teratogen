package fomalhaut

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
		default: Die("Bad octant");
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
						(float64(v) - 0.5) / (float64(u) / 0.5));
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