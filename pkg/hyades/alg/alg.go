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

// Cmp(a, b) returns -1, 0 or 1 if a is less than, equal to or greater than b,
// respectively. It will dereference pointers, convert interfaces to their
// underlying values, compare struct elements recursively and compare arrays
// and slices (arrays can be compared with slices) element by element.
// Functions and channels can not be compared.
func Cmp(lhs, rhs interface{}) (cmp int, ok bool) {
	return valueCmp(reflect.NewValue(lhs), reflect.NewValue(rhs))
}

func valueCmp(lhs, rhs reflect.Value) (cmp int, ok bool) {
	if lhs == nil {
		return -1, true
	}

	if rhs == nil {
		return 1, true
	}

	// Comparing arrays and slices is a special case where the values can be of
	// different types and still be compared. So check for it before checking
	// for type mismatch.
	if cmp, ok = arraySliceCmp(lhs, rhs); ok {
		return cmp, ok
	}

	if lhs.Type() != rhs.Type() {
		// Type mismatch.
		return
	}

	switch t := lhs.(type) {
	// Primitive types.
	case *reflect.BoolValue:
		v1, v2 := lhs.(*reflect.BoolValue).Get(), rhs.(*reflect.BoolValue).Get()
		if !v1 && v2 {
			return -1, true
		} else if v1 == v2 {
			return 0, true
		} else {
			return 1, true
		}
	case *reflect.FloatValue:
		v1, v2 := lhs.(*reflect.FloatValue).Get(), rhs.(*reflect.FloatValue).Get()
		if v1 < v2 {
			return -1, true
		} else if v1 == v2 {
			return 0, true
		} else {
			return 1, true
		}
	case *reflect.IntValue:
		v1, v2 := lhs.(*reflect.IntValue).Get(), rhs.(*reflect.IntValue).Get()
		if v1 < v2 {
			return -1, true
		} else if v1 == v2 {
			return 0, true
		} else {
			return 1, true
		}
	case *reflect.UintValue:
		v1, v2 := lhs.(*reflect.UintValue).Get(), rhs.(*reflect.UintValue).Get()
		if v1 < v2 {
			return -1, true
		} else if v1 == v2 {
			return 0, true
		} else {
			return 1, true
		}
	case *reflect.StringValue:
		v1, v2 := lhs.(*reflect.StringValue).Get(), rhs.(*reflect.StringValue).Get()
		if v1 < v2 {
			return -1, true
		} else if v1 == v2 {
			return 0, true
		} else {
			return 1, true
		}

	// Indirect types
	case *reflect.PtrValue:
		// Compare the values the pointers point to if the arguments are pointers.
		return valueCmp(
			lhs.(*reflect.PtrValue).Elem(),
			rhs.(*reflect.PtrValue).Elem())
	case *reflect.InterfaceValue:
		// Compare the values inside the interface.
		return Cmp(
			lhs.(*reflect.InterfaceValue).Elem(),
			rhs.(*reflect.InterfaceValue).Elem())

	// Non-slice, non-array composite values
	case *reflect.StructValue:
		return structCmp(lhs.(*reflect.StructValue), rhs.(*reflect.StructValue))
	case *reflect.MapValue:
		// TODO: Sketch: Extract keys, sort keys, make keys + values array, cmp arrays.
		panic("Map Cmp not implemented.")
	}

	return 0, false
}

func arraySliceCmp(lhs, rhs reflect.Value) (cmp int, ok bool) {
	seq1, ok1 := lhs.(reflect.ArrayOrSliceValue)
	seq2, ok2 := rhs.(reflect.ArrayOrSliceValue)
	if !ok1 || !ok2 {
		// Both aren't arrays or slices.
		return
	}

	if seq1.Type().(reflect.ArrayOrSliceType).Elem() != seq1.Type().(reflect.ArrayOrSliceType).Elem() {
		// Element type mismatch.
		return
	}

	if seq1.Len() != seq2.Len() {
		return Cmp(seq1.Len(), seq2.Len())
	}

	for i := 0; i < seq1.Len(); i++ {
		cmp, ok = valueCmp(seq1.Elem(i), seq2.Elem(i))
		if !ok {
			return 0, false
		}
		if cmp != 0 {
			return cmp, true
		}
	}

	return 0, true
}

func structCmp(lhs, rhs *reflect.StructValue) (cmt int, ok bool) {
	if lhs.NumField() != rhs.NumField() {
		return 0, false
	}

	t1 := lhs.Type().(*reflect.StructType)
	t2 := rhs.Type().(*reflect.StructType)
	for i := 0; i < lhs.NumField(); i++ {
		if t1.Field(i).Name != t2.Field(i).Name {
			return 0, false
		}
		if t1.Field(i).Type != t2.Field(i).Type {
			return 0, false
		}

		subCmp, subOk := valueCmp(lhs.Field(i), rhs.Field(i))
		if !subOk {
			return 0, false
		}
		if subCmp != 0 {
			return subCmp, true
		}
	}

	return 0, true
}
