package teratogen

import (
	"math"
	"testing"
	"testing/quick"
)

func normRollValid(m int16) bool {
	max := int(m)
	a := NormRoll(max)

	if max < 0 {
		max = 0
	}

	return -max <= a && a <= max
}

func TestNormRoll(t *testing.T) {
	if err := quick.Check(normRollValid, nil); err != nil {
		t.Error(err)
	}

	hist := map[int]int{
		-2: 0,
		-1: 0,
		0:  0,
		1:  0,
		2:  0}

	for i := 0; i < 10000; i++ {
		hist[NormRoll(2)] += 1
	}
	if hist[-2] > hist[-1] ||
		hist[-2] > hist[1] ||
		hist[2] > hist[-1] ||
		hist[2] > hist[1] ||
		hist[-1] > hist[0] ||
		hist[1] > hist[0] {
		t.Errorf("Bad histogram shape: %v", hist)
	}
}

func assertEqual(t *testing.T, expt, actual float64) {
	diff := math.Fabs(actual - expt)
	if diff > 0.01 {
		t.Errorf("Expected %v, got %v", expt, actual)
	}
}

func TestUnits(t *testing.T) {
	assertEqual(t, 64.0, ScaleToVolume(0))
	assertEqual(t, 64.0, ScaleToMass(0, 0))
	assertEqual(t, 1.587, ScaleToHeight(0))
}
