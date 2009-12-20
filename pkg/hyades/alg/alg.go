package alg

import (
	"container/vector"
	"sort"
)

// Ternary expression replacement.
func IfElse(exp bool, a interface{}, b interface{}) interface{} {
	if exp {
		return a
	}
	return b
}

type sortVec struct {
	data   *vector.Vector
	isLess func(i, j interface{}) bool
}

func (self *sortVec) Len() int { return self.data.Len() }
func (self *sortVec) Less(i, j int) bool {
	return self.isLess(self.data.At(i), self.data.At(j))
}
func (self *sortVec) Swap(i, j int) { self.data.Swap(i, j) }

// Return a new channel where the the first numItems from the previous channel
// are sorted using the sort predicate.
func PredicateSort(isLess func(i, j interface{}) bool, items *vector.Vector) {
	sortable := &sortVec{items, isLess}
	sort.Sort(sortable)
}
