package main

import (
	"hyades/dbg"
	"hyades/geom"
)

type Entity struct {
	Icon
	guid       Guid
	Name       string
	pos        geom.Pt2I
	parentId   Guid
	siblingId  Guid
	childId    Guid
	class      EntityClass
	isObstacle bool

	prop     map[string]interface{}
	hideProp map[string]bool
}

func NewEntity(guid Guid) (result *Entity) {
	result = new(Entity)
	result.prop = make(map[string]interface{})
	result.hideProp = make(map[string]bool)
	result.guid = guid
	return
}

func (self *Entity) IsObstacle() bool { return self.isObstacle }

func (self *Entity) GetPos() geom.Pt2I {
	parent := self.GetParent()
	if parent != nil {
		return parent.GetPos()
	}
	return self.pos
}

func (self *Entity) GetGuid() Guid { return self.guid }

func (self *Entity) GetClass() EntityClass { return self.class }

func (self *Entity) GetName() string { return self.Name }

func (self *Entity) MoveAbs(pos geom.Pt2I) { self.pos = pos }

func (self *Entity) Move(vec geom.Vec2I) { self.pos = self.pos.Plus(vec) }

func (self *Entity) GetParent() *Entity { return GetWorld().GetEntity(self.parentId) }

func (self *Entity) SetParent(e *Entity) { self.parentId = e.GetGuid() }

func (self *Entity) GetChild() *Entity { return GetWorld().GetEntity(self.childId) }

func (self *Entity) SetChild(e *Entity) { self.childId = e.GetGuid() }

// GetSibling return the next sibling of the entity, or nil if there are none.
func (self *Entity) GetSibling() *Entity { return GetWorld().GetEntity(self.siblingId) }

func (self *Entity) SetSibling(e *Entity) { self.siblingId = e.GetGuid() }

func (self *Entity) iterateChildren(c chan<- *Entity) {
	node := self.GetChild()
	for node != nil {
		c <- node
		for i := range node.Contents() {
			c <- i
		}
		node = node.GetSibling()
	}
}

func (self *Entity) Contents() <-chan *Entity {
	c := make(chan *Entity)
	go self.iterateChildren(c)
	return c
}

func (self *Entity) InsertSelf(parent *Entity) {
	self.RemoveSelf()
	if parent.GetChild() != nil {
		self.siblingId = parent.GetChild().GetGuid()
	}
	parent.SetChild(self)
}

func (self *Entity) RemoveSelf() {
	parent := self.GetParent()
	self.parentId = *new(Guid)
	if parent != nil {
		if parent.GetChild().GetGuid() == self.GetGuid() {
			parent.SetChild(self.GetSibling())
		} else {
			node := parent.GetChild()
			for {
				if node.GetSibling() == nil {
					dbg.Die("RemoveSelf: Entity not found among its siblings.")

				}
				if node.GetSibling().GetGuid() == self.GetGuid() {
					node.SetSibling(self.GetSibling())
					break
				}
				node = node.GetSibling()
			}
		}
	}
	self.siblingId = *new(Guid)
}

func (self *Entity) Set(name string, value interface{}) *Entity {
	self.hideProp[name] = false, false
	// TODO: Check that only valid value types pass.
	self.prop[name] = value
	return self
}

func (self *Entity) Get(name string) interface{} {
	_, hidden := self.hideProp[name]
	if hidden {
		return nil
	}
	ret, ok := self.prop[name]
	if ok {
		return ret
	}
	parent := self.PropParent()
	if parent != nil {
		return parent.Get(name)
	}
	return nil
}

func (self *Entity) Has(name string) bool { return self.Get(name) != nil }

func (self *Entity) Hide(name string) *Entity {
	self.hideProp[name] = true
	return self
}

func (self *Entity) Clear(name string) *Entity {
	self.hideProp[name] = false, false
	self.prop[name] = nil, false
	return self
}

func (self *Entity) PropParent() *Entity {
	// TODO
	return nil
}
