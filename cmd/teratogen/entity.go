package main

import (
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
	// GetChild return the first child of the entity, or nil if there are none.
	GetChild() Entity
	GetSibling() Entity
	Contents() <-chan Entity
}


type EntityBase struct {
	Icon
	guid      Guid
	Name      string
	pos       geom.Pt2I
	parentId  Guid
	siblingId Guid
	childId   Guid
	class     EntityClass
}

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

func (self *EntityBase) GetChild() Entity { return GetWorld().GetEntity(self.childId) }

// GetSibling return the next sibling of the entity, or nil if there are none.
func (self *EntityBase) GetSibling() Entity { return GetWorld().GetEntity(self.siblingId) }

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
