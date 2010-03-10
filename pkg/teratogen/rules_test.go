package teratogen

import (
	"math"
	"testing"
)

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
