package num

import (
	"math"
	"testing"
	"testing/quick"
)

func invSqrtError(x float64) float64 { return math.Fabs(InvSqrt(x) - 1.0/math.Sqrt(x)) }

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
