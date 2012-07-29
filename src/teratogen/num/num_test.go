/* num_test.go

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
	"testing"
	"testing/quick"
)

func invSqrtError(x float64) float64 { return math.Abs(InvSqrt(x) - 1.0/math.Sqrt(x)) }

const invSqrtTolerance = 0.01

func TestInvSqrt(t *testing.T) {
	invSqrtTest := func(x float64) bool {
		err := invSqrtError(x)
		if err > invSqrtTolerance {
			return false
		}
		return true
	}

	if err := quick.Check(invSqrtTest, nil); err != nil {
		t.Error(err)
	}
}
