package mem

import (
	"hyades/dbg"
	"reflect"
)

func TypeName(obj interface{}) string {
	val := reflect.Indirect(reflect.NewValue(obj))
	typ := val.Type()
	name := typ.Name()
	return name
}

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
	typ := reflect.Typeof(example)
	name := TypeName(example)
	dbg.Assert(name != "", "Unnamed type")
	self.typenames[name] = typ
}

// Create a blank object of a registered type.
func (self *BlankObjectFactory) Make(name string) interface{} {
	typ, ok := self.typenames[name]
	dbg.Assert(ok, "Unknown typename %v", name)
	return blankCopyOfType(typ).Interface()
}

// Make a null object of the same type as the parameter.
func BlankCopy(obj interface{}) interface{} {
	return blankCopyOfType(reflect.Typeof(obj)).Interface()
}

func blankCopyOfType(typ reflect.Type) reflect.Value {
	ptr, isPointer := typ.(*reflect.PtrType)
	if isPointer {
		// If the value is a pointer, make a new pointer value that
		// points to a new empty inner value.
		result := reflect.MakeZero(ptr).(*reflect.PtrValue)
		result.PointTo(reflect.MakeZero(ptr.Elem()))
		return result
	} else {
		// Otherwise just copy the value straight up.
		return reflect.MakeZero(typ)
	}
	panic("blankCopyFromValue")
}
