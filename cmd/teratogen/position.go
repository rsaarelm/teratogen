package main

import (
	"hyades/entity"
	"hyades/geom"
	"io"
)

const PosComponent = entity.ComponentFamily("position")


type posTemplate struct{}

func PosTemplate() *posTemplate { return new(posTemplate) }

func (self *posTemplate) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return self
}

func (self *posTemplate) MakeComponent(manager *entity.Manager, guid entity.Id) {
	GetManager().Handler(PosComponent).Add(guid, new(Position))
}


type Position struct {
	pos geom.Pt2I
}

func (self *Position) MoveAbs(pos geom.Pt2I) { self.pos = pos }

func (self *Position) Move(vec geom.Vec2I) { self.pos = self.pos.Plus(vec) }

func (self *Position) Serialize(out io.Writer) {
	entity.GobSerialize(out, self)
}

func (self *Position) Deserialize(in io.Reader) {
	entity.GobDeserialize(in, self)
}

func PosComp(guid entity.Id) *Position {
	if posComp := GetManager().Handler(PosComponent).Get(guid); posComp != nil {
		return posComp.(*Position)
	}
	return nil
}

func GetPos(guid entity.Id) (pos geom.Pt2I, ok bool) {
	if position := PosComp(guid); position != nil {
		return position.pos, true
	}

	ok = false
	return
}

// SetPos sets the entity's position if the entity has a position component.
// Returns false if the entity has no position component.
func SetPos(guid entity.Id, pos geom.Pt2I) bool {
	posComp := GetManager().Handler(PosComponent).Get(guid)
	if posComp != nil {
		posComp.(*Position).pos = pos
		return true
	}
	return false
}

// GetParentPosOrPos returns the GetParentPosOrPos of the container entity of
// entity guid if one exists and both entity guid and the container have a
// position component. Otherwise it works like GetPos.
func GetParentPosOrPos(guid entity.Id) (pos geom.Pt2I, ok bool) {
	if guid == entity.NilId {
		ok = false
		return
	}

	parentPos, parentOk := GetParentPosOrPos(GetParent(guid))
	if parentOk {
		pos = parentPos
		ok = true
		return
	}

	return GetPos(guid)
}
