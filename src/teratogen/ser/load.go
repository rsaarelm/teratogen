// load.go
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
	"unsafe"
)

func Load(input io.Reader) (topValue interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint("Load: ", e))
		}
	}()
	lo := newLoader(input)

	topValue = lo.loadSingle()
	for _, ok := lo.nextUnprocessed(); ok; _, ok = lo.nextUnprocessed() {
		lo.loadSingle()
	}

	lo.remapStalePointers()

	return
}

type loader struct {
	base
	// List stale pointers
	stalePointers []uintptr
	input         io.Reader
}

func newLoader(input io.Reader) (result *loader) {
	result = &loader{stalePointers: []uintptr{},
		input: input}

	result.processedObjects = make(map[uintptr]interface{})
	result.seenObjects = make(map[uintptr]interface{})
	return
}

func (lo *loader) Visit(value ...interface{}) {
	for _, v := range value {
		lo.visitSingle(v)
	}
}

func (lo *loader) Input() io.Reader { return lo.input }

func (lo *loader) Output() io.Writer { return nil }

func (lo *loader) TagPointer(obj interface{}) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("TagPointer called with non-pointer value %s", v))
	}
	// Stale pointers are used as IDs for the fresh objects, which we'll have
	// a full lookup of once all objects have been successfully deserialized.
	// Data pointers are first written as stale pointers, keeping a list of
	// all pointers *to* those pointers. When the stale to fresh pointer
	// lookup is complete, we will walk through the stale pointer list and
	// replace every one of them with fresh pointers.
	var stalePtr uintptr
	gobLoad(&stalePtr, lo.input)
	targetPtr := (*uintptr)(unsafe.Pointer(v.Pointer()))
	*targetPtr = stalePtr
	lo.stalePointers = append(lo.stalePointers, (uintptr)(v.Pointer()))
	lo.seenObjects[stalePtr] = nil
}

func (lo *loader) StoreGob(obj interface{}) {
	gobLoad(obj, lo.input)
}

func (lo *loader) visitSingle(obj interface{}) {
	switch obj.(type) {
	case *bool, *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16,
		*uint32, *uint64, *float32, *float64, *complex64, *complex128, *string:
		gobLoad(obj, lo.input)
	default:
		panic("Unhandled value type")
	}
}

func (lo *loader) remapStalePointers() {
	for _, ptr := range lo.stalePointers {
		targetPtr := (*uintptr)(unsafe.Pointer(ptr))
		fresh, ok := lo.processedObjects[*targetPtr]
		if !ok {
			panic(fmt.Sprintf("Unmapped stale pointer %s", ptr))
		}
		*targetPtr = reflect.ValueOf(fresh).Pointer()
	}
	lo.stalePointers = []uintptr{}
}

func (lo *loader) loadSingle() interface{} {
	var oldPtr uintptr
	var typeName string
	gobLoad(&oldPtr, lo.input)
	gobLoad(&typeName, lo.input)
	result := newInstance(typeName)
	err := result.Serialize(lo)
	if err != nil {
		panic(err)
	}
	lo.processedObjects[oldPtr] = result
	return result
}

func gobLoad(value interface{}, input io.Reader) {
	gobber := gob.NewDecoder(input)
	err := gobber.Decode(value)
	if err != nil {
		panic(err)
	}
}
