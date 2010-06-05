package alg

import (
	"exp/iterable"
)

// BK-Trees allow efficient search of nearby items using a fixed metric, such
// as correct spellings of a misspelled word based on Levenshtein distance.
type BkTree struct {
	edges map[int]*BkTree
	value interface{}
}

type BkMetric func(a, b interface{}) int

func NewBkTree(value interface{}) *BkTree { return &BkTree{make(map[int]*BkTree), value} }

// Add adds a new value to a BK-tree. The metric function must be same for all
// Add operations to the same tree, or the tree will end up in an invalid
// state.
func (self *BkTree) Add(metric BkMetric, value interface{}) {
	edge := metric(self.value, value)
	if node, ok := self.edges[edge]; ok {
		node.Add(metric, value)
	} else {
		self.edges[edge] = NewBkTree(value)
	}
}

func (self *BkTree) queryIter(c chan<- interface{}, maxDist int, metric BkMetric, value interface{}) {
	dist := metric(self.value, value)
	for i := dist - maxDist; i <= dist+maxDist; i++ {
		if i == 0 {
			c <- self.value
		}
		if node, ok := self.edges[i]; ok {
			node.queryIter(c, maxDist, metric, value)
		}
	}
}

// Query iterates the values in the BK-tree that are within maxDist of the
// given value. The metric function given must be the same which was used to
// build the tree.
func (self *BkTree) Query(maxDist int, metric BkMetric, value interface{}) iterable.Iterable {
	return iterable.Func(func(c chan<- interface{}) {
		self.queryIter(c, maxDist, metric, value)
		close(c)
	})
}
