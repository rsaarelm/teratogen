package mem

import (
	. "hyades/common"
	"reflect"
)

// Create an identifier for an object from its memory address.
func ObjId(obj interface{}) uintptr	{ return reflect.NewValue(obj).(*reflect.PtrValue).Get() }


// An object for keeping track of objects based on their memory addresses.
// Maintains reference counts of the object (updated by the user), and drops
// objects from the table when the count goes to 0.
type ObjLookup struct {
	lut		map[uintptr]interface{}
	objCount	map[uintptr]int
}

func NewObjLookup() (result *ObjLookup) {
	result = new(ObjLookup)
	result.lut = make(map[uintptr]interface{})
	result.objCount = make(map[uintptr]int)
	return
}

func (self *ObjLookup) GetObj(id uintptr) (result interface{}, found bool) {
	if obj, ok := self.lut[id]; ok {
		return obj, ok
	}
	return nil, false
}

func (self *ObjLookup) ContainsObj(obj interface{}) bool {
	if _, ok := self.lut[ObjId(obj)]; ok {
		return true
	}
	return false
}

// Increment references to a specific object. If there were no previous
// references, the object's id is added to the lookup table. Returns the id
// for the object.
func (self *ObjLookup) IncrObj(obj interface{}) uintptr {
	id := ObjId(obj)
	if count, ok := self.objCount[id]; ok {
		self.objCount[id] = count + 1
	} else {
		self.lut[id] = obj
		self.objCount[id] = 1
	}

	return id
}

// Decrements references to a spefic object. If the references go to zero,
// removes the object from the lookup. Decrementing a non-indexed object does
// nothing.
func (self *ObjLookup) DecrObj(obj interface{}) {
	id := ObjId(obj)
	if count, ok := self.objCount[id]; ok {
		if count-1 < 1 {
			self.RemoveObj(obj)
		} else {
			self.objCount[id] = count - 1
		}
	}	// if object not indexed, do nothing.
}

func (self *ObjLookup) RemoveObj(obj interface{}) {
	id := ObjId(obj)
	if _, ok := self.objCount[id]; ok {
		// Remove the object if found.
		self.lut[id] = obj, false
		self.objCount[id] = 0, false
	}	// if object not indexed, do nothing.
}

func (self *ObjLookup) iterate(c chan<- interface{}) {
	for _, val := range self.lut {
		c <- val
	}
	close(c)
}

func (self *ObjLookup) Iter() <-chan interface{} {
	c := make(chan interface{})
	go self.iterate(c)
	return c
}

func (self *ObjLookup) Len() int	{ return len(self.lut) }

// A factory object that matches the typenames to the type values of
// registered object types and is able to manufacture these objects given the
// string typename.
type BlankObjectFactory struct {
	typenames map[string]reflect.Type
}

func NewBlankObjectFactory() (result *BlankObjectFactory) {
	result = new(BlankObjectFactory)
	result.typenames = make(map[string]reflect.Type)
	return
}

// Register an object type in the factory.
func (self *BlankObjectFactory) Register(example interface{}) {
	val := reflect.Indirect(reflect.NewValue(example))
	typ := val.Type()
	name := typ.Name()
	Assert(name != "", "Unnamed type")
	self.typenames[name] = typ
}

// Create a blank object of a registered type.
func (self *BlankObjectFactory) Make(name string) interface{} {
	typ, ok := self.typenames[name]
	Assert(ok, "Unknown typename %v", name)
	return reflect.MakeZero(typ).Interface()
}
