package teratogen

import (
	"exp/iterable"
	"hyades/entity"
	"hyades/geom"
)

const WeaponComponent = entity.ComponentFamily("weapon")

type Weapon struct {
	Name  string
	Power float64
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

func (self *Weapon) Attack(wielder, target entity.Id, successDegree float64) {
	// SuccessDegree scales damage from 1/2 to full, critical hits double the damage.
	damage := self.Power/2 + successDegree*(self.Power/2)
	if successDegree == 1.0 {
		damage = self.Power * 2
	}

	GetCreature(target).Damage(
		target, wielder,
		GetPos(wielder), damage, BluntDamage)
}

func GetHitPos(origin, target geom.Pt2I) (hitPos geom.Pt2I) {
	for o := range iterable.Drop(geom.HexLine(origin, target), 1).Iter() {
		hitPos = o.(geom.Pt2I)
		if !IsOpen(hitPos) {
			break
		}
	}
	return
}

func Shoot(attackerId entity.Id, target geom.Pt2I) (endsMove bool) {
	dummyWeapon := &Weapon{"dummyGun", 1.0, 10}

	hitPos := GetHitPos(GetPos(attackerId), target)
	Fx().Shoot(attackerId, hitPos)

	DamagePos(hitPos, GetPos(attackerId), dummyWeapon.Power, PiercingDamage, attackerId)

	return true
}

func Attack(attackerId, targetId entity.Id) {
	dummyWeapon := &Weapon{"dummy", 1.0, 1}
	dummyWeapon.Attack(attackerId, targetId, 0.5)
}
