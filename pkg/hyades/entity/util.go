package entity

import (
	"container/vector"
	"exp/iterable"
	"hyades/mem"
	"io"
)

// uniqueComponentTable iterates through the components of a handler and gives
// each component with a unique memory address a consecutive index starting
// from 0. Return the corresponding list of unique values.
func uniqueComponentTable(handler Handler) (ids map[uintptr]int, values []interface{}) {
	ids = make(map[uintptr]int)
	vals := new(vector.Vector)
	i := 0
	for o := range handler.EntityComponents().Iter() {
		comp := o.(*IdComponent).Component
		id := mem.ObjId(comp)
		if _, ok := ids[id]; !ok {
			ids[id] = i
			vals.Push(comp)
			i++
		}
	}
	values = vals.Data()
	return
}

// SerializeHandlerComponents is an utility function that iterates through the
// components of a handler, and serializes components using a Serialize method
// which they must have.
func SerializeHandlerComponents(out io.Writer, handler Handler) {
	// Since the same entity may occur in multiple points, and we don't want to
	// flatten things, make a list of the unique entities with indices.
	uniqIds, uniqs := uniqueComponentTable(handler)

	// Serialize the actual component data, only saving each unique component
	// once.
	mem.WriteFixed(out, int32(len(uniqs)))
	for _, o := range uniqs {
		o.(mem.Serializable).Serialize(out)
	}

	// Write the entity ids and the indices to the unique component table.
	components := iterable.Data(handler.EntityComponents())
	mem.WriteFixed(out, int32(len(components)))
	for _, o := range components {
		pair := o.(*IdComponent)
		id := uniqIds[mem.ObjId(pair.Component)]

		mem.WriteFixed(out, int64(pair.Entity))
		mem.WriteFixed(out, int32(id))
	}
}

// DeserializeHandlerComponents reads a sequence of components saved with
// SerializeHandlerComponents from the input stream, constructs them using the
// provided newComponent function, deserializes them with a Deserialize method
// which they must have, and adds them to the handler.
func DeserializeHandlerComponents(in io.Reader, handler Handler, newComponent func() interface{}) {
	// Deserialize each unique component.
	nUniqs := int(mem.ReadInt32(in))
	uniqs := make([]interface{}, nUniqs)
	for i := 0; i < nUniqs; i++ {
		uniqs[i] = newComponent()
		uniqs[i].(mem.Serializable).Deserialize(in)
	}

	nComponents := int(mem.ReadInt32(in))
	for i := 0; i < nComponents; i++ {
		guid := Id(mem.ReadInt64(in))
		uniqId := int(mem.ReadInt32(in))
		handler.Add(guid, uniqs[uniqId])
	}
}
