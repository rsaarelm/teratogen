package geom

import (
	"exp/iterable"
	"hyades/alg"
	"testing"
)

func pointsEqual(o1, o2 interface{}) bool {
	p1, ok1 := o1.(Pt2I)
	p2, ok2 := o2.(Pt2I)
	if !ok1 || !ok2 {
		return false
	}
	if !p1.Equals(p2) {
		return false
	}
	return true
}

func testLine(t *testing.T, p1, p2 Pt2I, coords ...) {
	ptVec := alg.UnpackEllipsis(coords)
	pts1 := make([]interface{}, len(ptVec)/2)
	for i := 0; i < len(ptVec); i += 2 {
		pts1[i/2] = Pt2I{ptVec[i].(int), ptVec[i+1].(int)}
	}
	pts2 := iterable.Data(Line(p1, p2))
	if !alg.ArraysEqual(pointsEqual, pts1, pts2) {
		t.Errorf("Expected %#v, got %#v", pts1, pts2)
	}
}

func TestLine(t *testing.T) {
	testLine(t, Pt2I{0, 0}, Pt2I{0, 0},
		0, 0)
	testLine(t, Pt2I{0, 0}, Pt2I{2, 0},
		0, 0,
		1, 0,
		2, 0)

	testLine(t, Pt2I{0, 0}, Pt2I{-2, 0},
		0, 0,
		-1, 0,
		-2, 0)

	testLine(t, Pt2I{0, 0}, Pt2I{0, 2},
		0, 0,
		0, 1,
		0, 2)

	testLine(t, Pt2I{0, 0}, Pt2I{0, -2},
		0, 0,
		0, -1,
		0, -2)

	testLine(t, Pt2I{2, 0}, Pt2I{5, 7},
		2, 0,
		2, 1,
		3, 2,
		3, 3,
		4, 4,
		4, 5,
		5, 6,
		5, 7)
}
