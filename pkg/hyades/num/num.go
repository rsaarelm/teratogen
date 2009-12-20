package num

import (
	"math"
)

func IntMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Float64Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func Round(x float64) float64 { return math.Floor(x + 0.5) }

func Iabs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func Fsignum(x float64) float64 {
	switch {
	case x < 0.0:
		return -1.0
	case x > 0.0:
		return 1.0
	}
	return 0.0
}

func Isignum(x int) int {
	switch {
	case x < 0:
		return -1
	case x > 0:
		return 1
	}
	return 0
}

// Base-2 logarithm.
func Log2(x float64) float64 { return math.Log(x) / math.Log(2.0) }

// Deterministic noise in [-1.0, 1.0). From Hugo Elias,
// http://freespace.virgin.net/hugo.elias/models/m_perlin.htm
func Noise(seed int) float64 {
	seed = (seed << 13) ^ seed
	return (1.0 -
		float64((seed*(seed*seed*15731+789221)+1376312589)&0x7fffffff)/
			1073741824.0)
}

// Fracf returns the fractional part of f.
func Fracf(f float64) (frac float64) {
	_, frac = math.Modf(f)
	return
}

// Linear interpolation between a and b using x = [0, 1].
func Lerp(a, b float64, x float64) float64 { return a + (b-a)*x }
