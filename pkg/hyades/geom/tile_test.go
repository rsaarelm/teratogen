package geom

import (
	"testing"
	"testing/quick"
)

func TestArrayHex(t *testing.T) {
	// It does get breaky if you give it large enough numbers to cause an
	// overflow. Int16 params keep the cap on the size.
	test := func(x, y int16) bool {
		p0 := Pt2I{int(x), int(y)}
		p1 := Array2Hex(Hex2Array(p0))
		p2 := Hex2Array(Array2Hex(p0))
		return p0.Equals(p1) && p1.Equals(p2)
	}

	if err := quick.Check(test, nil); err != nil {
		t.Error(err)
	}
}

func TestHexDist(t *testing.T) {
	distMatrix := [][]int{
		[]int{4, 3, 2, 2, 2},
		[]int{3, 2, 1, 1, 2},
		[]int{2, 1, 0, 1, 2},
		[]int{2, 1, 1, 2, 3},
		[]int{2, 2, 2, 3, 4},
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
