package geom

import (
	"testing"
	"testing/quick"
)

func TestPlaneHex(t *testing.T) {
	test := func(x, y int16) bool {
		p0 := Pt2I{int(x), int(y)}
		px, py := HexToPlane(p0)
		p1 := PlaneToHex(px, py)
		return p0.Equals(p1)
	}

	if err := quick.Check(test, nil); err != nil {
		t.Error(err)
	}
}

func TestHexDist(t *testing.T) {
	distMatrix := [][]int{
		[]int{2, 2, 2, 3, 4},
		[]int{2, 1, 1, 2, 3},
		[]int{2, 1, 0, 1, 2},
		[]int{3, 2, 1, 1, 2},
		[]int{4, 3, 2, 2, 2},
	}
	min := -len(distMatrix) / 2
	max := len(distMatrix) / 2

	for y := min; y <= max; y++ {
		for x := min; x <= max; x++ {
			pt := Vec2I{x, y}
			expect := distMatrix[y-min][x-min]
			dist := HexDist(Origin, Origin.Plus(pt))
			if dist != expect {
				t.Errorf("Dist to %v should be %v, was %v\n", pt, expect, dist)
			}
		}
	}
}
