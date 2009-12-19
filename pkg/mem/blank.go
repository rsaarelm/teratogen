package mem

import (
	"fmt"
	"hyades/dbg"
	"reflect"
)

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
	dbg.Assert(name != "", "Unnamed type")
	self.typenames[name] = typ
}

// Create a blank object of a registered type.
func (self *BlankObjectFactory) Make(name string) interface{} {
	typ, ok := self.typenames[name]
	dbg.Assert(ok, "Unknown typename %v", name)
	return reflect.MakeZero(typ).Interface()
}

// Make a null object of the same type as the parameter.
func BlankCopy(obj interface{}) interface{} {
	ptr, isPointer := reflect.NewValue(obj).(*reflect.PtrValue)
	if isPointer {
		// If the value is a pointer, make a new pointer value that
		// points to a new empty inner value.
		val := reflect.MakeZero(ptr.Elem().Type())
		wrap := reflect.MakeZero(ptr.Type()).(*reflect.PtrValue)
		wrap.PointTo(val)
		return wrap.Interface()
	} else {
		// Otherwise just copy the value straight up.
		typ := reflect.Typeof(obj)
		fmt.Println(typ)
		return reflect.MakeZero(typ).Interface()
	}
	panic("BlankCopy")
}
