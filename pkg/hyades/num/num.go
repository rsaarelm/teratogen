// The num package provides mathematical, numerical and random number
// generation utilities.

package num

import (
	"math"
)

func Imax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Imin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Fmax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func Fmin(a, b float64) float64 {
	if a < b {
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


// Fracf returns the fractional part of f.
func Fracf(f float64) (frac float64) {
	_, frac = math.Modf(f)
	return
}

// Lerp does linear interpolation between a and b using parameter 0 <= x <= 1.
func Lerp(a, b float64, x float64) float64 { return a*(1-x) + b*x }

// CosInterp does cosine interpolation between a and b using parameter 0 <= x
// <= 1.
func CosInterp(a, b float64, x float64) float64 {
	f := (1 - math.Cos(x*math.Pi)) * 0.5
	return Lerp(a, b, f)
}

func Clamp(min, max, x float64) float64 {
	if x < min {
		return min
	} else if x > max {
		return max
	}
	return x
}
