package fomalhaut

type Set interface {
	Add(item interface{});
	Remove(item interface{});
	Contains(item interface{}) bool;
	Len() int;
	// TODO: Iter() <-chan interface{};
	Items() []interface{};
}

type MapSet struct {
	items *ObjLookup;
}

func NewMapSet() (result *MapSet) {
	result = new(MapSet);
	result.items = NewObjLookup();

	return;
}

func (self *MapSet)Add(item interface{}) {
	// Also tracks count of times added, though we don't care about that.
	self.items.IncrObj(item);
}

func (self *MapSet)Remove(item interface{}) {
	self.items.RemoveObj(item);
}

func (self *MapSet)Contains(item interface{}) bool {
	return self.items.ContainsObj(item);
}

func (self *MapSet)Len() int { return self.items.Len(); }