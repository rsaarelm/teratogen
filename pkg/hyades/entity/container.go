package entity

import (
	"exp/iterable"
	"hyades/alg"
	"hyades/mem"
	"io"
)

// Default component container type. This can be used directly for handlers
// that don't do anything other than contain a set of homogenous
// non-interface-typed components.
type Container struct {
	components         map[Id]interface{}
	componentPrototype interface{}
}

// Init initializes a new container handler. Prototype is an instance of the
// concrete component type this container holds.
func (self *Container) Init(prototype interface{}) {
	self.components = make(map[Id]interface{})
	self.componentPrototype = prototype
}

func (self *Container) Add(guid Id, component interface{}) {
	self.components[guid] = component
}

func (self *Container) Remove(guid Id) { self.components[guid] = nil, false }

func (self *Container) Get(guid Id) interface{} {
	if result, ok := self.components[guid]; ok {
		return result
	}
	return nil
}

func (self *Container) Serialize(out io.Writer) {
	SerializeHandlerComponents(out, self)
}

func (self *Container) Deserialize(in io.Reader) {
	DeserializeHandlerComponents(in, self, mem.BlankCopier(self.componentPrototype))
}

func (self *Container) EntityComponents() iterable.Iterable {
	return alg.IterFunc(func(c chan<- interface{}) {
		for id, comp := range self.components {
			c <- &IdComponent{id, comp}
		}
		close(c)
	})
}
