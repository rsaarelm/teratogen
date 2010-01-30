// A component-based entity system.

package entity

import (
	"container/vector"
	"exp/iterable"
	"io"
	"hyades/alg"
	"hyades/dbg"
	"hyades/mem"
)

// Entities are nothing but GUID values.
type Id int64

const NilId = Id(0)

type ComponentFamily string

// Assemblages specify the initial component sets of new entities.
type Assemblage map[ComponentFamily]interface{}

// The manager holds all the component handlers and creates new entities.
type Manager struct {
	nextGuid     int64
	handlers     map[ComponentFamily]Handler
	liveEntities map[Id]bool
}

func NewManager() (result *Manager) {
	result = new(Manager)
	result.handlers = make(map[ComponentFamily]Handler)
	result.liveEntities = make(map[Id]bool)
	return
}

// NewEntity returns a new unique entity identifier.
func (self *Manager) NewEntity() (result Id) {
	self.nextGuid++
	result = Id(self.nextGuid)
	self.liveEntities[result] = true
	return
}

// Handler returns the component handler for the given component family.
func (self *Manager) Handler(family ComponentFamily) Handler {
	return self.handlers[family]
}

// SetHandler sets the component handler for the given component family.
func (self *Manager) SetHandler(family ComponentFamily, handler Handler) {
	self.handlers[family] = handler
}

// BuildEntity creates a new entity, gives it the component values in the
// component families specified by the assemblage and returns the entity
// value.
func (self *Manager) BuildEntity(assemblage Assemblage) (result Id) {
	result = self.NewEntity()
	for family, component := range assemblage {
		self.Handler(family).Add(result, component)
	}
	return
}

// RemoveEntity removes the entity from the Manager. It is removed from all
// component systems it has a component in.
func (self *Manager) RemoveEntity(entity Id) {
	for _, handler := range self.handlers {
		handler.Remove(entity)
	}
	self.liveEntities[entity] = false, false
}

// Entities iterates through live entities in the manager. An entity is live
// if it has been created by BuildEntity and has not yet been removed with
// RemoveEntity.
func (self *Manager) Entities() iterable.Iterable {
	return alg.IterFunc(func(c chan<- interface{}) {
		for entity, _ := range self.liveEntities {
			c <- entity
		}
	})
}

// Serialize stores the entity state of the manager into a stream.
func (self *Manager) Serialize(out io.Writer) {
	mem.WriteFixed(out, self.nextGuid)

	mem.WriteFixed(out, int32(len(self.liveEntities)))
	for ent, _ := range self.liveEntities {
		mem.WriteFixed(out, int64(ent))
	}

	mem.WriteFixed(out, int32(len(self.handlers)))
	for family, handler := range self.handlers {
		mem.WriteString(out, string(family))
		handler.Serialize(out)
	}
}

// Deserialize loads the entity state from a stream. Note that all the Handler
// families that are expected to come up from the stream must already be
// initialized with proper instances in the Manager before Deserialize is
// called. Otherwise the Manager can't deserialize individual Handlers.
func (self *Manager) Deserialize(in io.Reader) {
	self.nextGuid = mem.ReadInt64(in)

	nLiveEntities := int(mem.ReadInt32(in))
	for i := 0; i < nLiveEntities; i++ {
		ent := Id(mem.ReadInt64(in))
		self.liveEntities[ent] = true
	}

	nHandlers := int(mem.ReadInt32(in))
	for i := 0; i < nHandlers; i++ {
		family := ComponentFamily(mem.ReadString(in))
		handler, ok := self.handlers[family]
		dbg.Assert(ok, "Handler for family '%v' not ready for deserialization.", family)
		handler.Deserialize(in)
	}
}

type IdComponent struct {
	Entity    Id
	Component interface{}
}

// IdComponent2Id maps IdComponent pointer to its Entity field. Handy with
// iterable.Map.
func IdComponent2Id(obj interface{}) interface{} {
	return obj.(*IdComponent).Entity
}

// IdComponent2Component maps IdComponent pointer to its Component field.
// Handy with iterable.Map.
func IdComponent2Component(obj interface{}) interface{} {
	return obj.(*IdComponent).Component
}

// Handler handles the entire collection of one type of component for
// the game state.
type Handler interface {
	// Add adds a component for the given entity.
	Add(guid Id, component interface{})
	// Remove removes this type of component from the given entity.
	Remove(guid Id)
	// Get looks up this type of component for the given entity, return nil if
	// component wasn't found.
	Get(guid Id) interface{}
	// Serialize saves this set of components to a stream.
	Serialize(out io.Writer)
	// Deserialize initializes a new handler from a stream.
	Deserialize(in io.Reader)
	// EntityComponents iterates through the entity, component pairs in this
	// handler as IdComponent values.
	EntityComponents() iterable.Iterable
}

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
