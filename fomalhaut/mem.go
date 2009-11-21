package fomalhaut

import "reflect"

// Create an identifier for an object from its memory address.
func ObjId(obj interface{}) uintptr {
	return reflect.NewValue(obj).(*reflect.PtrValue).Get();
}


// An object for keeping track of objects based on their memory addresses.
// Maintains reference counts of the object (updated by the user), and drops
// objects from the table when the count goes to 0.
type ObjLookup struct {
	lut map[uintptr] interface{};
	objCount map[uintptr] int;
}

func NewObjLookup() (result *ObjLookup) {
	result = new(ObjLookup);
	result.lut = make(map[uintptr] interface{});
	result.objCount = make(map[uintptr] int);
	return;
}

func (self *ObjLookup)GetObj(id uintptr) (result interface{}, found bool) {
	if obj, ok := self.lut[id]; ok {
		return obj, ok;
	}
	return nil, false;
}

// Increment references to a specific object. If there were no previous
// references, the object's id is added to the lookup table. Returns the id
// for the object.
func (self *ObjLookup)IncrObj(obj interface{}) uintptr {
	id := ObjId(obj);
	if count, ok := self.objCount[id]; ok {
		self.objCount[id] = count + 1;
	} else {
		self.lut[id] = obj;
		self.objCount[id] = 1;
	}

	return id;
}

// Decrements references to a spefic object. If the references go to zero,
// removes the object from the lookup. Decrementing a non-indexed object does
// nothing.
func (self *ObjLookup)DecrObj(obj interface{}) {
	id := ObjId(obj);
	if count, ok := self.objCount[id]; ok {
		if count - 1 < 1 {
			// No references,
			self.lut[id] = obj, false;
			self.objCount[id] = 0, false;
		} else {
			self.objCount[id] = count - 1;
		}
	} // if object not indexed, do nothing.
}

func (self *ObjLookup)Objects() (result []interface{}) {
	result = make([]interface{}, len(self.lut));
	i := 0;
	for _, val := range self.lut {
		result[i] = val;
		i++;
	}
	return result;
}