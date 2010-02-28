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
