package geom

import (
	"hyades/num"
	"log"
	"math"
)

// Recursive shadowcasting line-of-sight algorithm. Always starts from 0, 0.
// Function parameter isBlocked tells if a tile relative to the origin blocks
// sight, and function parameter outsideRadius is false for points within
// sight radius and true for points outside it. The algorithm will span a
// connected and convex area based on outsideRadius and yield the visible
// points in the returned channel.
func LineOfSight(isBlocked func(Vec2I) bool, outsideRadius func(Vec2I) bool) <-chan Vec2I {
	c := make(chan Vec2I)

	go func() {
		c <- Vec2I{0, 0}

		for octant := 0; octant < 8; octant++ {
			processOctant(
				c, isBlocked, outsideRadius, octant,
				0.0, 1.0, 1)
		}
		close(c)
	}()

	return c
}

func processOctant(c chan<- Vec2I, isBlocked func(Vec2I) bool, outsideRadius func(Vec2I) bool, octant int, startSlope float64, endSlope float64, u int) {
	if outsideRadius(Vec2I{u, 0}) {
		return
	}

	if endSlope <= startSlope {
		return
	}

	traversingObstacle := true

	for v, ev := int(num.Round(float64(u)*startSlope)), int(math.Ceil(float64(u)*endSlope)); v <= ev; v++ {
		var pos Vec2I
		switch octant {
		case 0:
			pos = Vec2I{v, -u}
		case 1:
			pos = Vec2I{u, -v}
		case 2:
			pos = Vec2I{u, v}
		case 3:
			pos = Vec2I{v, u}
		case 4:
			pos = Vec2I{-v, u}
		case 5:
			pos = Vec2I{-u, v}
		case 6:
			pos = Vec2I{-u, -v}
		case 7:
			pos = Vec2I{-v, -u}
		default:
			log.Crashf("Bad octant %v", octant)
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
					(float64(v)-0.5)/(float64(u)+0.5),
					u+1)
				traversingObstacle = true
			}
		} else {
			if traversingObstacle {
				// Risen above an obstacle.
				traversingObstacle = false
				if v > 0 {
					startSlope = num.Fmax(
						startSlope,
						(float64(v)-0.5)/(float64(u)-0.5))
				}
			}
		}

		if startSlope <= endSlope && !outsideRadius(pos) {
			c <- pos
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
			u+1)
	}
}

// Determine the 1/16th sector of a circle a point in the XY plane points
// towards. Sector 0 is clockwise from the y-axis, and subsequent sectors are
// clockwise from there. The origin point is handled in the same way as in
// math.Atan2.
func Hexadecant(x, y float64) int {
	const hexadecantWidth = math.Pi / 8.0
	radian := math.Atan2(x, -y)
	if radian < 0 {
		radian += 2.0 * math.Pi
	}
	return int(math.Floor(radian / hexadecantWidth))
}

func Vec2IToDir8(vec Vec2I) int {
	return ((Hexadecant(float64(vec.X), float64(vec.Y)) + 1) % 16) / 2
}

func Vec2IToDir4(vec Vec2I) int {
	return ((Hexadecant(float64(vec.X), float64(vec.Y)) + 2) % 16) / 4
}

// Vec2IToDir6 converts a vector to a hex direction when the vector is from a
// rectilinear coordinate system underlying the hex grid. Dual to Dir6ToVec.
func Vec2IToDir6(vec Vec2I) int {
	hexadecant := Hexadecant(float64(vec.X), float64(vec.Y))
	switch hexadecant {
	case 14, 15, 0:
		return 0
	case 1, 2:
		return 1
	case 3, 4, 5:
		return 2
	case 6, 7, 8:
		return 3
	case 9, 10:
		return 4
	case 11, 12, 13:
		return 5
	}
	log.Crashf("Bad hexadecant %v", hexadecant)
	return 0
}

func Dir8ToVec(dir int) (result Vec2I) {
	switch dir {
	case 0:
		return Vec2I{0, -1}
	case 1:
		return Vec2I{1, -1}
	case 2:
		return Vec2I{1, 0}
	case 3:
		return Vec2I{1, 1}
	case 4:
		return Vec2I{0, 1}
	case 5:
		return Vec2I{-1, 1}
	case 6:
		return Vec2I{-1, 0}
	case 7:
		return Vec2I{-1, -1}
	}
	log.Crashf("Invalid dir %v", dir)
	return
}

func PosAdjacent(p1, p2 Pt2I) bool {
	diff := p1.Minus(p2)
	x, y := num.Iabs(diff.X), num.Iabs(diff.Y)
	return x < 2 && y < 2 && x+y > 0
}

// Dir6ToVec converts a hex direction to a vector when using a hex map system
// where the hexes are superimposed on rectilinear coordinates.
func Dir6ToVec(dir int) (result Vec2I) {
	switch dir {
	case 0:
		return Vec2I{0, -1}
	case 1:
		return Vec2I{1, -1}
	case 2:
		return Vec2I{1, 0}
	case 3:
		return Vec2I{0, 1}
	case 4:
		return Vec2I{-1, 1}
	case 5:
		return Vec2I{-1, 0}
	}
	log.Crashf("Invalid dir %v", dir)
	return
}

// http://www-cs-students.stanford.edu/~amitp/Articles/HexLOS.html

// HexDist returns the hexagonal distance between two points.
func HexDist(p1, p2 Pt2I) int {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y

	if num.Isignum(dx) == num.Isignum(dy) {
		return num.Imax(num.Iabs(dx), num.Iabs(dy))
	}
	return num.Iabs(dx) + num.Iabs(dy)
}
