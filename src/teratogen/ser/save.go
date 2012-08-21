// save.go
//
// Copyright (C) 2012 Risto Saarelma
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package ser

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"reflect"
)

func Save(topValue interface{}, output io.Writer) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint("Save: ", e))
		}
	}()
	s := newSaver(output)

	s.saveSingle(topValue.(Serializable))

	for ptr, ok := s.nextUnprocessed(); ok; ptr, ok = s.nextUnprocessed() {
		s.saveSingle(s.seenObjects[ptr].(Serializable))
	}
	return
}

type saver struct {
	base
	output io.Writer
}

func newSaver(output io.Writer) (result *saver) {
	result = &saver{output: output}

	result.processedObjects = make(map[uintptr]interface{})
	result.seenObjects = make(map[uintptr]interface{})
	return
}

func (s *saver) Visit(value ...interface{}) {
	for _, v := range value {
		s.visitSingle(v)
	}
}

func (s *saver) Input() io.Reader { return nil }

func (s *saver) Output() io.Writer { return s.output }

func (s *saver) TagPointer(obj interface{}) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("TagPointer called with non-pointer value %s", v))
	}
	v = reflect.Indirect(v)
	gobSave(v.Pointer(), s.output)
	s.seenObjects[v.Pointer()] = v.Interface()
}

func (s *saver) visitSingle(obj interface{}) {
	switch obj.(type) {
	case *bool, *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16,
		*uint32, *uint64, *float32, *float64, *complex64, *complex128, *string:
		deref := reflect.Indirect(reflect.ValueOf(obj))
		gobSave(deref.Interface(), s.output)
	default:
		panic("Unhandled value type")
	}
}

func (s *saver) saveSingle(obj Serializable) {
	v := reflect.ValueOf(obj)
	gobSave(v.Pointer(), s.output)

	if name, ok := concreteTypeToName[reflect.TypeOf(obj).Elem()]; ok {
		gobSave(name, s.output)
	} else {
		panic(fmt.Sprintf("Serializing unregistered type %s", reflect.TypeOf(obj)))
	}

	err := obj.Serialize(s)
	if err != nil {
		panic(err)
	}

	s.processedObjects[v.Pointer()] = obj
}

func gobSave(value interface{}, output io.Writer) {
	gobber := gob.NewEncoder(output)
	err := gobber.Encode(value)
	if err != nil {
		panic(err)
	}
}
