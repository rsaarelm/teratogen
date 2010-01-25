// A component-based entity system.

package entity

import (
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
	return
}

// NewEntity returns a new unique entity identifier.
func (self *Manager) newEntity() Id {
	self.nextGuid++
	return Id(self.nextGuid)
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
	result = self.newEntity()
	self.liveEntities[result] = true
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

type EntityComponent struct {
	Entity    Id
	Component interface{}
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
	// handler as EntityComponent values.
	EntityComponents() iterable.Iterable
}
