package teratogen

import (
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
)

const PosComponent = entity.ComponentFamily("position")


type posTemplate struct{}

func PosTemplate() *posTemplate { return new(posTemplate) }

func (self *posTemplate) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return self
}

func (self *posTemplate) MakeComponent(manager *entity.Manager, guid entity.Id) {
	manager.Handler(PosComponent).Add(guid, new(Position))
}


// The position component.
type Position struct {
	pos geom.Pt2I

	// velocity is the entity's movement vector from it's last turn. Use it for
	// dodge bonuses, charge attacks etc.
	velocity geom.Vec2I

	// velocityStale indicates the velocity hasn't been updated recently and
	// should be cleared.
	velocityStale bool
}

// Pos returns the position in the position component.
func (self *Position) Pos() geom.Pt2I { return self.pos }

// Velocity returns the velocity in the position component.
func (self *Position) Velocity() geom.Vec2I { return self.velocity }

// MoveAbs sets the position in the position component.
func (self *Position) MoveAbs(pos geom.Pt2I) { self.pos = pos }

// Move adds the given vector to the position in the position component.
func (self *Position) Move(vec geom.Vec2I) {
	self.velocity = vec
	self.velocityStale = false
	self.pos = self.pos.Plus(vec)
}

func (self *Position) ClearStaleVelocity() {
	if self.velocityStale {
		self.velocity = geom.ZeroVec2I
	}
}

// MakeVelocityStale makes the current velocity "stale", to be set to zero on
// next time ClearStaleVelocity is called, but calling Move after calling
// MakeVelocityStale will remove the staleness.
func (self *Position) MakeVelocityStale() {
	self.velocityStale = true
}

func PosComp(id entity.Id) *Position {
	if posComp := GetManager().Handler(PosComponent).Get(id); posComp != nil {
		return posComp.(*Position)
	}
	return nil
}

func HasPosComp(guid entity.Id) bool { return PosComp(guid) != nil }

func GetPos(guid entity.Id) (result geom.Pt2I) {
	if position := PosComp(guid); position != nil {
		return position.pos
	}
	dbg.Die("%v has no position component.", guid)
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

// GetParentPosOrPos returns the position of the eldest container parent of an
// entity that has a position component. For entities that are not contained
// in a parent, it works like GetPos.
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

	if ok = HasPosComp(guid); ok {
		pos = GetPos(guid)
	}
	return
}

func TryMove(id entity.Id, vec geom.Vec2I) (success bool) {
	if crit := GetCreature(id); crit != nil {
		if crit.HasIntrinsic(IntrinsicImmobile) {
			// Immobiles can't ever move.
			return false
		}
	}

	if IsOpen(GetPos(id).Plus(vec)) {
		PosComp(id).Move(vec)
		return true
	}
	return false
}
