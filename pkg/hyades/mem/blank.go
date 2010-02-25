package mem

import (
	"gob"
	"hyades/dbg"
	"io"
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
	gobErr := IsGobSerializable(example)
	if gobErr != nil {
		dbg.Warn("Type %s isn't gob-serializable: %v", name, gobErr)
	}
	self.typenames[name] = typ
}

// Create a blank object of a registered type.
func (self *BlankObjectFactory) Make(name string) interface{} {
	typ, ok := self.typenames[name]
	dbg.Assert(ok, "Unknown typename %v", name)
	return blankCopyOfType(typ).Interface()
}

// Save an object's typename and the object itself to outstream. Use gob
// serialization on the object.
func (self *BlankObjectFactory) GobSave(out io.Writer, obj interface{}) {
	name := TypeName(obj)
	_, ok := self.typenames[name]
	dbg.Assert(ok, "Trying to save object of unrecognized type %s.", name)
	WriteString(out, name)
	enc := gob.NewEncoder(out)
	err := enc.Encode(obj)
	dbg.AssertNoError(err)
}

// Load a gob-serialized object preceded by its typename from the instream.
// The object's type must be registered in the factory so a blank copy can be
// made based on the typename.
func (self *BlankObjectFactory) GobLoad(in io.Reader) interface{} {
	name := ReadString(in)
	obj := self.Make(name)
	dec := gob.NewDecoder(in)
	err := dec.Decode(obj)
	dbg.AssertNoError(err)
	return obj
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
		result.PointTo(blankCopyOfType(ptr.Elem()))
		return result
	}
	// Otherwise just copy the value straight up.
	return reflect.MakeZero(typ)
}

// BlankCopier returns a function that creates blank copies of the type of the
// argument value.
func BlankCopier(obj interface{}) func() interface{} {
	return func() interface{} { return BlankCopy(obj) }
}
