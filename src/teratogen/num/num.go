/* num.go

   Copyright (C) 2012 Risto Saarelma

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

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

// AbsMod is a modulo operation where -12 modulo 10 is 8, not -2.
func AbsMod(x, modulo int) int {
	if x < 0 {
		return x%modulo + modulo
	}
	return x % modulo
}
