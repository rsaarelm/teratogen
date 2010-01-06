package geom

import (
	"exp/iterable"
	"hyades/alg"
	"hyades/num"
	"math"
)

// TraceRay iterates an unlimited number of consecutive Pt2I ray points
// starting from orig and moving towards vector [dx, dy]. The absolute
// magnitude of [dx, dy] is ignored, except if it is less than epsilon in
// which case [dx, dy] becomes [1, 0] (similar to calling math.Atan2(0.0,
// 0.0)).
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

// TraceLine iterates the consecutive Pt2I points along the line from p1 to p2
func Line(p1, p2 Pt2I) iterable.Iterable {
	ray := Ray(p1, float64(p2.X-p1.X), float64(p2.Y-p1.Y))
	nPoints := num.Imax(num.Iabs(p2.X-p1.X), num.Iabs(p2.Y-p1.Y)) + 1
	return iterable.Take(ray, nPoints)
}
