package geom

import (
	"exp/iterable"
	"hyades/alg"
	"hyades/num"
	"math"
)

// Ray iterates an unlimited number of consecutive Pt2I ray points starting
// from orig and moving towards vector [dx, dy]. The absolute magnitude of
// [dx, dy] is ignored, except if it is less than epsilon in which case [dx,
// dy] becomes [1, 0] (similar to calling math.Atan2(0.0, 0.0)).
func Ray(orig Pt2I, dx, dy float64) iterable.Iterable {
	const epsilon = 1e-10
	if dx*dx+dy*dy < epsilon {
		dx, dy = 1, 0
	}

	// Pick a scale such that the larger of dx, dy can be normalized to unit length.
	scale := num.Fmax(math.Fabs(dx), math.Fabs(dy))
	dx /= scale
	dy /= scale
	return alg.IterFunc(func(c chan<- interface{}) {
		x, y := float64(orig.X), float64(orig.Y)
		for {
			c <- Pt2I{int(num.Round(x)), int(num.Round(y))}
			x += dx
			y += dy
		}
	})
}

func HexRay(orig Pt2I, dx, dy float64) iterable.Iterable {
	const epsilon = 1e-10
	if dx*dx+dy*dy < epsilon {
		dx, dy = 1, 0
	}

	// Pick a scale such that the larger of dx, dy can be normalized to
	// one-half unit length. (Hex rays need more precision than square grid
	// rays.)
	scale := num.Fmax(math.Fabs(dx), math.Fabs(dy)) * 2
	dx /= scale
	dy /= scale
	prev := orig

	return alg.IterFunc(func(c chan<- interface{}) {
		x, y := HexToPlane(orig)
		// Send the starting point here as it'll otherwise be skipped in the
		// Equals prev test.
		c <- orig
		for {
			pt := PlaneToHex(x, y)
			if !pt.Equals(prev) {
				c <- pt
				prev = pt
			}
			x += dx
			y += dy
		}
	})
}

// Line iterates the consecutive Pt2I points along the line from p1 to p2.
func Line(p1, p2 Pt2I) iterable.Iterable {
	vec := p2.Minus(p1)
	ray := Ray(p1, float64(vec.X), float64(vec.Y))
	nPoints := num.Imax(num.Iabs(vec.X), num.Iabs(vec.Y)) + 1
	return iterable.Take(ray, nPoints)
}

func HexLine(p1, p2 Pt2I) iterable.Iterable {
	vec := p2.Minus(p1)
	dx, dy := HexToPlane(Pt2I{vec.X, vec.Y})
	ray := HexRay(p1, dx, dy)
	running := true
	whilePred := func(o interface{}) (result bool) {
		result = running
		running = !o.(Pt2I).Equals(p2)
		return
	}
	return iterable.TakeWhile(ray, whilePred)
}
