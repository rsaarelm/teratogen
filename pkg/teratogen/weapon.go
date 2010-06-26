package teratogen

import (
	"hyades/entity"
	"hyades/geom"
)

const WeaponComponent = entity.ComponentFamily("weapon")

type Weapon struct {
	Name  string
	Power int
	Range int
}

// Serve as template, prototype-style.

func (self *Weapon) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return c
}

func (self *Weapon) MakeComponent(manager *entity.Manager, guid entity.Id) {
	result := new(Weapon)
	*result = *self
	manager.Handler(WeaponComponent).Add(guid, result)
}

func (self *Weapon) CanAttack(wielder, target entity.Id) bool {
	// TODO: Return whether the target is in weapon's range and within clear
	// line of sight.
	pos0 := GetPos(wielder)
	pos1 := GetPos(target)
	if geom.HexDist(pos0, pos1) > self.Range {
		return false
	}

	// Attacks are only possible along hex axes.
	if _, ok := geom.Vec2IToDir6Exact(pos1.Minus(pos0)); !ok {
		return false
	}

	if _, ok := geom.HexLineOfSight(pos0, pos1, IsBlocked); !ok {
		return false
	}

	return true
}

func (self *Weapon) Attack(wielder, target entity.Id) {
	// TODO: Actual attack logic, move to damage if attack works.
}
