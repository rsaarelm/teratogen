package teratogen

import (
	"exp/iterable"
	"hyades/entity"
	"hyades/geom"
)

const (
	NoWeapon = iota
	WeaponFist
	WeaponBayonet
	WeaponClaw
	WeaponKick
	WeaponHorns
	WeaponJaws
	WeaponPistol
	WeaponRifle
	WeaponBile
	WeaponCrawl
	WeaponSpider
	WeaponCyclops
	WeaponSaw
	WeaponZap
	WeaponSmash
	WeaponNether
)

const WeaponComponent = entity.ComponentFamily("weapon")

type Weapon struct {
	Name  string
	Verb  string
	Power float64
	Range int
}

var weaponLookup = map[int]*Weapon{
	NoWeapon:      nil,
	WeaponFist:    &Weapon{"fist", "hit{sub.s}", 10, 1},
	WeaponBayonet: &Weapon{"bayonet", "hit{sub.s}", 20, 1},
	WeaponClaw:    &Weapon{"claw", "claw{sub.s}", 15, 1},
	WeaponKick:    &Weapon{"hooves", "kick{sub.s}", 15, 1},
	WeaponHorns:   &Weapon{"horns", "headbutt{sub.s}", 15, 1},
	WeaponJaws:    &Weapon{"bite", "bite{sub.s}", 15, 1},
	WeaponPistol:  &Weapon{"pistol", "shoot{sub.s}", 10, 7},
	WeaponRifle:   &Weapon{"rifle", "shoot{sub.s}", 24, 12},
	WeaponBile:    &Weapon{"bile", "vomit{sub.s} bile at", 19, 5},
	WeaponCrawl:   &Weapon{"touch", "{.section sub.you}touch{.or}touches{.end}", 10, 1},
	// TODO: Poison
	WeaponSpider:  &Weapon{"bite", "bite{sub.s}", 30, 1},
	WeaponCyclops: &Weapon{"psychic blast", "blast{sub.s}", 24, 7},
	WeaponSaw:     &Weapon{"chainsaw", "chainsaw{sub.s}", 35, 1},
	WeaponZap:     &Weapon{"electro-zapper", "zap{sub.s}", 15, 4},
	WeaponSmash:   &Weapon{"mighty smash", "hit{sub.s}", 40, 1},
	WeaponNether:  &Weapon{"nether ray", "exhale{sub.s}", 40, 7},
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

	// TODO: Attack effect as weapon data, not just this ad-hoc thing.
	if self.Range > 1 {
		Fx().Shoot(wielder, GetPos(target))
	}

	EMsg("{sub.Thename} %s {obj.thename}.\n", wielder, target, self.Verb)

	// TODO: Damage type from weapon.
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
	dummyWeapon := &Weapon{"dummyGun", "shoot{sub.s}", 24.0, 10}

	hitPos := GetHitPos(GetPos(attackerId), target)
	Fx().Shoot(attackerId, hitPos)

	EMsg("{sub.Thename} %s.\n", attackerId, entity.NilId, dummyWeapon.Verb)
	DamagePos(hitPos, GetPos(attackerId), dummyWeapon.Power, PiercingDamage, attackerId)

	return true
}

func Attack(attackerId, targetId entity.Id) {
	crit := GetCreature(attackerId)
	// XXX: Fixed the first weapon as preferred melee attack.
	if weapon := weaponLookup[crit.Attack1]; weapon != nil {
		// TODO: Get success level.
		weapon.Attack(attackerId, targetId, 0.5)
	}
}
