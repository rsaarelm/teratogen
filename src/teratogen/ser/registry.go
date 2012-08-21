// registry.go
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
	"fmt"
	"reflect"
)

// Copied stuff over from gob source. Basically need the exact same thing it
// does here, but gob won't expose the registry API in a reusable way.
// Simplified a bunch, since my system only expects to deal with named
// Serializable types.

var (
	nameToConcreteType = make(map[string]reflect.Type)
	concreteTypeToName = make(map[reflect.Type]string)
)

// RegisterName is like Register but uses the provided name rather than the
// type's default.
func RegisterName(name string, value Serializable) {
	if name == "" {
		// reserved for nil
		panic("attempt to register empty name")
	}
	rt := reflect.TypeOf(value).Elem()
	// Check for incompatible duplicates. The name must refer to the
	// same user type, and vice versa.
	if t, ok := nameToConcreteType[name]; ok && t != rt {
		panic(fmt.Sprintf("ser: registering duplicate types for %q: %s != %s", name, t, rt))
	}
	if n, ok := concreteTypeToName[rt]; ok && n != name {
		panic(fmt.Sprintf("ser: registering duplicate names for %s: %q != %q", rt, n, name))
	}

	nameToConcreteType[name] = rt
	concreteTypeToName[rt] = name
}

// Register records a type, identified by a value for that type, under its
// internal type name.
func Register(obj Serializable) {
	name := reflect.TypeOf(obj).Elem().Name()
	if name == "" {
		panic("Registering an unnamed type")
	}
	RegisterName(name, obj)
}

func newInstance(name string) Serializable {
	if t, ok := nameToConcreteType[name]; ok {
		resultVal := reflect.New(t)
		if result, castOk := (resultVal.Interface()).(Serializable); castOk {
			return result
		} else {
			panic(fmt.Sprintf("Couldn't cast type '%s' to Serializable", name))
		}
	}
	panic(fmt.Sprintf("Tried to create unregistered type '%s'", name))
}
