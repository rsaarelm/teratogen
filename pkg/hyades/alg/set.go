package alg

import (
	"container/vector"
	"hyades/mem"
)

type Set interface {
	Add(item interface{})
	Remove(item interface{})
	Contains(item interface{}) bool
	Len() int
	Iter() <-chan interface{}
}

// XXX: Element iteration order uses memory addresses of element objects, and
// is therefore undefined relative to the order of element insertion.
type MapSet struct {
	items *mem.ObjLookup
}

func NewMapSet() (result *MapSet) {
	result = new(MapSet)
	result.items = mem.NewObjLookup()

	return
}

func (self *MapSet) Add(item interface{}) {
	// Also tracks count of times added, though we don't care about that.
	self.items.IncrObj(item)
}

func (self *MapSet) Remove(item interface{}) { self.items.RemoveObj(item) }

func (self *MapSet) Contains(item interface{}) bool {
	return self.items.ContainsObj(item)
}

func (self *MapSet) iterate(c chan<- interface{}) {
	for i := range self.items.Iter() {
		c <- i
	}
	close(c)
}

func (self *MapSet) Iter() <-chan interface{} {
	c := make(chan interface{})
	go self.iterate(c)
	return c
}

func (self *MapSet) Len() int { return self.items.Len() }

// VecSet implements set with an inefficient linear lookup but with guaranteed
// stable element iteration order.
type VecSet vector.Vector

func NewVecSet() *VecSet {
	return (*VecSet)(new(vector.Vector))
}

func (self *VecSet) Add(item interface{}) {
	(*vector.Vector)(self).Push(item)
}

func (self *VecSet) find(item interface{}) (idx int, ok bool) {
	id := mem.ObjId(item)
	for i, o := range *self {
		if mem.ObjId(o) == id {
			ok = true
			idx = i
			return
		}
	}
	return
}

func (self *VecSet) Remove(item interface{}) {
	if idx, ok := self.find(item); ok {
		(*vector.Vector)(self).Delete(idx)
	}
}

func (self *VecSet) Contains(item interface{}) bool {
	_, ok := self.find(item)
	return ok
}

func (self *VecSet) Len() int {
	return (*vector.Vector)(self).Len()
}

func (self *VecSet) Iter() <-chan interface{} {
	c := make(chan interface{})
	go func() {
		for _, o := range *self {
			c <- o
		}
		close(c)
	}()
	return c
}
