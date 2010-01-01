// Miscellaneous low-level program logic utilities.

package alg

import (
	"container/vector"
	"exp/iterable"
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

// Return a new channel where the the first numItems from the previous channel
// are sorted using the sort predicate.
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

type IterFunc func(c chan<- interface{})

func (self IterFunc) Iter() <-chan interface{} {
	c := make(chan interface{})
	go self(c)
	return c
}

// ReverseIter returns an Iterable that returns the elements if iter in
// reverse order.
func ReverseIter(iter iterable.Iterable) iterable.Iterable {
	data := iterable.Data(iter)
	return IterFunc(func(c chan<- interface{}) {
		for i := len(data) - 1; i >= 0; i-- {
			c <- data[i]
		}
		close(c)
	})
}

func emptyIteration(c chan<- interface{}) { close(c) }

// EmptyIter returns an Iterable that yields nothing. Useful in situations
// where an interface requires that an Iterable is presented.
func EmptyIter() iterable.Iterable { return IterFunc(emptyIteration) }
