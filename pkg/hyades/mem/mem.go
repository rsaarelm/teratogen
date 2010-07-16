// Data storage, data format and memory operations.

package mem

import (
	"bytes"
	"fmt"
	"gob"
	"os"
	"reflect"
)

// Create an identifier for an object from its memory address.
func ObjId(obj interface{}) uintptr { return reflect.NewValue(obj).(*reflect.PtrValue).Get() }

// Return nil if obj can be serialized with gob, an error describing the
// problem if it can't.
func IsGobSerializable(obj interface{}) os.Error {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	return enc.Encode(obj)
}

func AssignFields(obj reflect.Value, values map[string]interface{}) os.Error {
	obj = reflect.Indirect(obj)
	fields, ok := obj.(*reflect.StructValue)
	if !ok {
		return os.NewError(fmt.Sprintf("AssignFields: Value %v isn't a struct.", obj))
	}

	for k, v := range values {
		field := fields.FieldByName(k)
		if field == nil {
			return os.NewError(fmt.Sprintf("AssignFields: Field '%s' not in struct.", k))
		}
		val := reflect.NewValue(v)
		if field.Type() != val.Type() {
			return os.NewError(fmt.Sprintf(
				"AssignFields: Type mismatch for '%s', %v in data for %v in struct.",
				k, val.Type(), field.Type()))
		}
		field.SetValue(reflect.NewValue(v))
	}

	return nil
}
