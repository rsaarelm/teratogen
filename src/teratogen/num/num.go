// num.go
//
// Copyright (C) 2012 Risto Saarelma
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package num provides miscellaneous numerical utilities.
package num

import (
	"image"
	"math"
)

func Round(x float64) float64 { return math.Floor(x + 0.5) }

// Base-2 logarithm.
func Log2(x float64) float64 { return math.Log(x) / math.Log(2.0) }

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

// InvSqrt approximates the inverse square root of x very quickly.
func InvSqrt(x float64) float64 {
	const sqrtMagic64 = 0x5FE6EC85E7DE30DA

	// Initial guess.
	tmp := math.Float64frombits(sqrtMagic64 - math.Float64bits(x)>>1)

	return tmp * (1.5 - 0.5*x*tmp*tmp)
}

func NumberOfSetBitsU64(x uint64) (result int) {
	for i := uint64(0); i < 64; i++ {
		if x&(1<<i) != 0 {
			result++
		}
	}
	return
}

// AbsMod is a modulo operation which maps negative numbers to [0, modulo).
func AbsMod(x, modulo int) int {
	if x < 0 {
		return x%modulo + modulo
	}
	return x % modulo
}

// Noise generates pseudorandom noise values. From Hugo Elias,
// http://freespace.virgin.net/hugo.elias/models/m_perlin.htm
func Noise(seed int) float64 {
	seed = (seed << 13) ^ seed
	return (1.0 -
		float64((seed*(seed*seed*15731+789221)+1376312589)&0x7fffffff)/1073741824.0)
}

// SignI return -1 for negative, 0 for zero and 1 for positive integer values.
func SignI(x int) int {
	if x < 0 {
		return -1
	} else if x > 0 {
		return 1
	}
	return 0
}

func AbsI(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// BresenhamLine traces a line from p1 to p2 using Bresenham's line drawing
// algorithm and calls the callback function for each point along the line.
func BresenhamLine(cb func(image.Point), p1, p2 image.Point) {
	data := line(p1, p2)
	err := 0
	pos := p1
	for i := 0; i <= data.DMajor; i++ {
		cb(pos)
		err += data.DMinor
		if (2 * err) >= data.DMajor {
			pos = pos.Add(data.SideAxis)
			err -= data.DMajor
		}
		pos = pos.Add(data.MainAxis)
	}
}

type lineData struct {
	DMajor   int
	DMinor   int
	MainAxis image.Point
	SideAxis image.Point
}

func line(p1, p2 image.Point) (result lineData) {
	dx, dy := p2.X-p1.X, p2.Y-p1.Y
	signX, signY := SignI(dx), SignI(dy)
	absDx, absDy := AbsI(dx), AbsI(dy)
	if absDx > absDy {
		result.DMajor = absDx
		result.DMinor = absDy
		result.MainAxis = image.Pt(signX, 0)
		result.SideAxis = image.Pt(0, signY)
	} else {
		result.DMajor = absDy
		result.DMinor = absDx
		result.MainAxis = image.Pt(0, signY)
		result.SideAxis = image.Pt(signX, 0)
	}
	return
}
