package main

import (
	"hyades/geom"
)

type Item struct {
	Icon
	guid	Guid
	Name	string
	pos	geom.Pt2I
	class	EntityClass
}

func (self *Item) IsObstacle() bool	{ return false }

func (self *Item) GetPos() geom.Pt2I	{ return self.pos }

func (self *Item) GetGuid() Guid	{ return self.guid }

func (self *Item) GetClass() EntityClass	{ return self.class }

func (self *Item) GetName() string	{ return self.Name }

func (self *Item) MoveAbs(pos geom.Pt2I)	{ self.pos = pos }

func (self *Item) Move(vec geom.Vec2I)	{ self.pos = self.pos.Plus(vec) }
