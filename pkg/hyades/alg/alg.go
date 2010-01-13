// Miscellaneous low-level program logic utilities.

package alg

import (
	"container/vector"
	"reflect"
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

// PredicateSort sorts the values in the given vector using the given ordering
// predicate.
func PredicateSort(isLess func(i, j interface{}) bool, items *vector.Vector) {
	sortable := &sortVec{items, isLess}
	sort.Sort(sortable)
}

// UnpackEllipsis converts a ... parameter into an array of interface{}
// values, each being one parameter in the ... list.
func UnpackEllipsis(a ...) (result []interface{}) {
	v := reflect.NewValue(a).(*reflect.StructValue)
	result = make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		result[i] = v.Field(i).Interface()
	}
	return
}

// ChanData reads the output from a channel into an array.
func ChanData(in <-chan interface{}) (result []interface{}) {
	vec := new(vector.Vector)
	for x := range in {
		vec.Push(x)
	}
	result = make([]interface{}, vec.Len())
	for i, _ := range result {
		result[i] = vec.At(i)
	}
	return
}

// ArraysEqual returns whether two arrays have the same length and have all
// corresponding elements satisfy the equality predicate.
func ArraysEqual(eqPred func(e1, e2 interface{}) bool, a1, a2 []interface{}) bool {
	if len(a1) != len(a2) {
		return false
	}
	for i, _ := range a1 {
		if !eqPred(a1[i], a2[i]) {
			return false
		}
	}
	return true
}
