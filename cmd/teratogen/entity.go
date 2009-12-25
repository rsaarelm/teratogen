package main

import (
	"hyades/dbg"
	"hyades/geom"
)

// IMPORTANT: With the current savegame implementation, all structs that
// implement Entity and go into the Entity store in World, MUST be
// gob-serializable. That means no field values that are interfaces, maps,
// channels or funcs.
type Entity interface {
	Drawable
	// TODO: Entity-common stuff.
	IsObstacle() bool
	GetPos() geom.Pt2I
	GetGuid() Guid
	MoveAbs(pos geom.Pt2I)
	Move(vec geom.Vec2I)
	GetName() string
	GetClass() EntityClass

	GetParent() Entity
	SetParent(e Entity)
	// GetChild return the first child of the entity, or nil if there are none.
	GetChild() Entity
	SetChild(e Entity)
	GetSibling() Entity
	SetSibling(e Entity)
	Contents() <-chan Entity
	InsertSelf(parent Entity)
	RemoveSelf()

	// Properties
	Set(name string, value interface{}) (self Entity)
	Get(name string) interface{}
	Has(name string) bool
	Hide(name string) (self Entity)
	Clear(name string) (self Entity)
	PropParent() Entity
}


type EntityBase struct {
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

func NewEntity(guid Guid) (result *EntityBase) {
	result = new(EntityBase)
	result.prop = make(map[string]interface{})
	result.hideProp = make(map[string]bool)
	result.guid = guid
	return
}

func (self *EntityBase) IsObstacle() bool { return self.isObstacle }

func (self *EntityBase) GetPos() geom.Pt2I {
	parent := self.GetParent()
	if parent != nil {
		return parent.GetPos()
	}
	return self.pos
}

func (self *EntityBase) GetGuid() Guid { return self.guid }

func (self *EntityBase) GetClass() EntityClass {
	return self.class
}

func (self *EntityBase) GetName() string { return self.Name }

func (self *EntityBase) MoveAbs(pos geom.Pt2I) {
	self.pos = pos
}

func (self *EntityBase) Move(vec geom.Vec2I) { self.pos = self.pos.Plus(vec) }

func (self *EntityBase) GetParent() Entity { return GetWorld().GetEntity(self.parentId) }

func (self *EntityBase) SetParent(e Entity) { self.parentId = e.GetGuid() }

func (self *EntityBase) GetChild() Entity { return GetWorld().GetEntity(self.childId) }

func (self *EntityBase) SetChild(e Entity) { self.childId = e.GetGuid() }

// GetSibling return the next sibling of the entity, or nil if there are none.
func (self *EntityBase) GetSibling() Entity { return GetWorld().GetEntity(self.siblingId) }

func (self *EntityBase) SetSibling(e Entity) { self.siblingId = e.GetGuid() }

func (self *EntityBase) iterateChildren(c chan<- Entity) {
	node := self.GetChild()
	for node != nil {
		c <- node
		for i := range node.Contents() {
			c <- i
		}
		node = node.GetSibling()
	}
}

func (self *EntityBase) Contents() <-chan Entity {
	c := make(chan Entity)
	go self.iterateChildren(c)
	return c
}

func (self *EntityBase) InsertSelf(parent Entity) {
	self.RemoveSelf()
	if parent.GetChild() != nil {
		self.siblingId = parent.GetChild().GetGuid()
	}
	parent.SetChild(self)
}

func (self *EntityBase) RemoveSelf() {
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

func (self *EntityBase) Set(name string, value interface{}) Entity {
	self.hideProp[name] = false, false
	// TODO: Check that only valid value types pass.
	self.prop[name] = value
	return self
}

func (self *EntityBase) Get(name string) interface{} {
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

func (self *EntityBase) Has(name string) bool { return self.Get(name) != nil }

func (self *EntityBase) Hide(name string) Entity {
	self.hideProp[name] = true
	return self
}

func (self *EntityBase) Clear(name string) Entity {
	self.hideProp[name] = false, false
	self.prop[name] = nil, false
	return self
}

func (self *EntityBase) PropParent() Entity {
	// TODO
	return nil
}
