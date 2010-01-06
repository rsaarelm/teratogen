package alg

import (
	"exp/iterable"
	"math"
)

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

// IterMin returns the value from the iteration that is smallest according to
// the isLess function.
func IterMin(iter iterable.Iterable, measure func(e1 interface{}) float64) (result interface{}, ok bool) {
	ok = false
	minVal := float64(math.MaxFloat64)
	for e := range iter.Iter() {
		m := measure(e)
		if !ok || m < minVal {
			ok = true
			result = e
			minVal = m
		}
	}
	return
}
