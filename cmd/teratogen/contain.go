package main

import (
	"exp/iterable"
	"hyades/alg"
	"hyades/entity"
)

const ContainComponent = entity.ComponentFamily("contain")

// GetContain gets the containment relation. Lhs is the container and rhs
// values are the immediate containees (transitive containment, being in a
// container in a container, won't show up in the top relation).
func GetContain() *entity.Relation {
	return GetManager().Handler(ContainComponent).(*entity.Relation)
}

// TODO: Turn inventory methods from Blob component methods into plain
// functions on entity guids.

func (self *Blob) GetParent() *Blob {
	if id, ok := GetContain().GetLhs(self.GetGuid()); ok {
		return GetBlobs().Get(id).(*Blob)
	}
	return nil
}

// GetTopParent gets the greatest grandparent of the entity.
func (self *Blob) GetTopParent() (result *Blob) {
	for p := self.GetParent(); p != nil; p = p.GetParent() {
		result = p
	}
	return
}

func (self *Blob) SetParent(e *Blob) {
	if e != nil {
		GetContain().AddPair(e.GetGuid(), self.GetGuid())
	} else {
		GetContain().RemoveWithRhs(self.GetGuid())
	}
}

// RecursiveContents iterates through all children and grandchildren of the
// entity.
func (self *Blob) RecursiveContents() iterable.Iterable {
	return alg.IterFunc(func(c chan<- interface{}) {
		for o := range self.Contents().Iter() {
			c <- o
			for q := range o.(*Blob).RecursiveContents().Iter() {
				c <- q
			}
		}
		close(c)
	})
}

// Contents iterates through the children but not the grandchildren of the
// entity.
func (self *Blob) Contents() iterable.Iterable {
	return iterable.Map(GetContain().IterRhs(self.GetGuid()), id2Blob)
}

func (self *Blob) HasContents() bool {
	_, ok := GetContain().GetRhs(self.GetGuid())
	return ok
}

func (self *Blob) InsertSelf(parent *Blob) {
	self.RemoveSelf()
	GetContain().AddPair(parent.GetGuid(), self.GetGuid())
}

func (self *Blob) RemoveSelf() {
	if parent := self.GetTopParent(); parent != nil {
		self.MoveAbs(parent.GetPos())
	}
	GetContain().RemoveWithRhs(self.GetGuid())
}
