package gamelib

import "math"

func IntMax(a, b int) int {
	if a > b {
		return a;
	}
	return b;
}

func Float64Max(a, b float64) float64 {
	if a > b {
		return a;
	}
	return b;
}

func Round(x float64) float64 { return math.Floor(x + 0.5); }

func Iabs(x int) int {
	if x < 0 { return -x; }
	return x;
}

func Fsignum(x float64) float64 {
	switch {
	case x < 0.0: return -1.0;
	case x > 0.0: return 1.0;
	}
	return 0.0;
}

func Isignum(x int) int {
	switch {
	case x < 0: return -1;
	case x > 0: return 1;
	}
	return 0;
}

// Base-2 logarithm.
func Log2(x float64) float64 { return math.Log(x) / math.Log(2.0); }