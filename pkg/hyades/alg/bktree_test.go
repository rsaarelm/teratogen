package alg

import (
	"math"
	"testing"
)

func metric(a, b interface{}) int {
	return int(math.Fabs(float64(a.(int)) - float64(b.(int))))
}

func TestBkTree(t *testing.T) {
	tree := NewBkTree(0)
	tree.Add(metric, 10)
	tree.Add(metric, 12)
	tree.Add(metric, 15)
	tree.Add(metric, 22)
	// XXX: Hacky brittle test routine
	count := 0
	for x := range tree.Query(4, metric, 18).Iter() {
		i := x.(int)
		if i == 15 || i == 22 {
			count++
			// ok
		} else {
			t.Errorf("Bad query result %s", i)
		}
	}
	if count != 2 {
		t.Errorf("Missing query items.")
	}
}
