package fomalhaut

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