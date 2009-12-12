package gamelib

type Set interface {
	Add(item interface{})
	Remove(item interface{})
	Contains(item interface{}) bool
	Len() int
	Iter() <-chan interface{}
}

type MapSet struct {
	items *ObjLookup
}

func NewMapSet() (result *MapSet) {
	result = new(MapSet)
	result.items = NewObjLookup()

	return
}

func (self *MapSet) Add(item interface{}) {
	// Also tracks count of times added, though we don't care about that.
	self.items.IncrObj(item)
}

func (self *MapSet) Remove(item interface{})	{ self.items.RemoveObj(item) }

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

func (self *MapSet) Len() int	{ return self.items.Len() }
