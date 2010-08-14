package teratogen

import (
	"exp/iterable"
	"hyades/entity"
	"hyades/geom"
	"hyades/num"
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
	WeaponGaze
	WeaponPsiBlast
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
	Flags int64
}

var weaponLookup = map[int]*Weapon{
	NoWeapon:       nil,
	WeaponFist:     &Weapon{"fist", "hit{sub.s}", 10, 1, 0},
	WeaponBayonet:  &Weapon{"bayonet", "hit{sub.s}", 20, 1, 0},
	WeaponClaw:     &Weapon{"claw", "claw{sub.s}", 15, 1, 0},
	WeaponKick:     &Weapon{"hooves", "kick{sub.s}", 15, 1, 0},
	WeaponHorns:    &Weapon{"horns", "headbutt{sub.s}", 15, 1, 0},
	WeaponJaws:     &Weapon{"bite", "bite{sub.s}", 15, 1, 0},
	WeaponPistol:   &Weapon{"pistol", "shoot{sub.s}", 30, 7, WeaponUsesAmmo},
	WeaponRifle:    &Weapon{"rifle", "shoot{sub.s}", 24, 12, WeaponUsesAmmo},
	WeaponBile:     &Weapon{"bile", "vomit{sub.s} bile at", 19, 5, 0},
	WeaponCrawl:    &Weapon{"touch", "{.section sub.you}touch{.or}touches{.end}", 10, 1, 0},
	WeaponSpider:   &Weapon{"bite", "bite{sub.s}", 30, 1, 0},     // TODO: Poison
	WeaponGaze:     &Weapon{"gaze", "gazes{sub.s} at", 24, 7, 0}, // TODO: Confuse
	WeaponPsiBlast: &Weapon{"psychic blast", "blast{sub.s}", 24, 7, 0},
	WeaponSaw:      &Weapon{"chainsaw", "chainsaw{sub.s}", 35, 1, 0},
	WeaponZap:      &Weapon{"electro-zapper", "zap{sub.s}", 15, 4, 0}, // TODO: Stun
	WeaponSmash:    &Weapon{"mighty smash", "hit{sub.s}", 40, 1, 0},
	WeaponNether:   &Weapon{"nether ray", "exhale{sub.s}", 40, 7, 0},
}

const (
	WeaponUsesAmmo = 1 << iota
)

// Serve as template, prototype-style.

func (self *Weapon) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return c
}

func (self *Weapon) MakeComponent(manager *entity.Manager, guid entity.Id) {
	result := new(Weapon)
	*result = *self
	manager.Handler(WeaponComponent).Add(guid, result)
}

func (self *Weapon) HasFlag(flag int64) bool {
	return (self.Flags & flag) != 0
}

func (self *Weapon) CanAttack(wielder entity.Id, pos geom.Pt2I) bool {
	pos0 := GetPos(wielder)
	if geom.HexDist(pos0, pos) > self.Range {
		return false
	}

	// Attacks are only possible along hex axes.
	if _, ok := geom.Vec2IToDir6Exact(pos.Minus(pos0)); !ok {
		return false
	}

	if _, ok := geom.HexLineOfSight(pos0, pos, BlocksRanged); !ok {
		return false
	}

	// XXX: Doesn't check whether the line of attack is physically blocked by,
	// say, another entity. Friendly fire fun.

	return true
}

// Attack attacks a position with the weapon. Does not check whether the
// weapon allows attacking that pos. Some weapons may hit targets in between
// the attacker and the target pos.
func (self *Weapon) Attack(wielder entity.Id, pos geom.Pt2I, attackBonus float64) {
	isRangedAttack := self.Range > 1

	if !self.ExpendAmmo(wielder) {
		EMsg("{sub.Thename} {sub.is} out of ammo.\n", wielder, entity.NilId)
		return
	}

	var successDegree float64

	for o := range iterable.Drop(geom.HexLine(GetPos(wielder), pos), 1).Iter() {
		hitPos := o.(geom.Pt2I)
		successDegree = self.checkHits(hitPos, attackBonus)
		if successDegree >= 0 {
			pos = hitPos
			break
		}
	}

	if isRangedAttack {
		// TODO: Attack effect as weapon data, not just this ad-hoc thing.
		Fx().Shoot(wielder, pos)
	}

	target := CreatureAt(pos)
	if target == entity.NilId {
		// No creature at final pos, attacking empty air.
		EMsg("{sub.Thename} %s.\n", wielder, target, self.Verb)
		return
	}

	if successDegree < 0 {
		// Missed a creature.
		EMsg("{sub.Thename} {.section sub.you}miss{.or}misses{.end} {obj.thename}.\n", wielder, target)
		return
	}

	isCriticalHit := successDegree == 1.0

	// SuccessDegree scales damage from 1/8 to full.
	damage := num.Lerp(self.Power/8, self.Power, successDegree)
	if isCriticalHit {
		// TODO: Special effect for critical hit. Just amping the damage is
		// bad, as it throws off the attrition. Some kind of weapon-specific
		// status damage like stunning or a crippling or bleeding wound would
		// be nice though.
	}

	EMsg("{sub.Thename} %s {obj.thename}.\n", wielder, target, self.Verb)

	// TODO: Damage type from weapon.
	GetCreature(target).Damage(
		target, wielder,
		GetPos(wielder), damage, BluntDamage)
}

// checkHits returns whether the weapon's attack hits something in the given
// cell. The result is a success roll. Nonnegative values mean a degree of
// hit, negative values mean a miss. It doesn't specify what is hit. It might
// be a terrain tile or an entity. Entities can evade attacks, so the result
// for open tiles containing an entity target can be somewhat random.
func (self *Weapon) checkHits(pos geom.Pt2I, attackBonus float64) float64 {
	if !IsOpenTerrain(pos) {
		return 1.0
	}

	for o := range EntitiesAt(pos).Iter() {
		id := o.(entity.Id)
		if BlocksMovement(id) {
			success := ContestRoll(attackBonus - DefenseSkill(id))
			if success >= 0 {
				return success
			}
		}
	}

	return -1.0
}

// ExpendAmmo checks whether the weapon consumes ammo. If it does and the
// wielder is an ammo-tracking entity, subtract ammo from the wielder. If the
// wielder is out of ammo, return false to indicate that the attack is not
// possible. If there was ammo left or the weapon doesn't use ammo, return
// true.
func (self *Weapon) ExpendAmmo(wielder entity.Id) (attackPossible bool) {
	if self.HasFlag(WeaponUsesAmmo) {
		if inv := GetInventory(wielder); inv != nil {
			if inv.Ammo == 0 {
				return false
			}
			inv.Ammo--
			return true
		}
	}
	return true
}

func Shoot(attackerId entity.Id, target geom.Pt2I) (endsMove bool) {
	crit := GetCreature(attackerId)

	// XXX: Ugly hardcoding for always using second weapon for shooting.
	if weapon := weaponLookup[crit.Attack2]; weapon != nil {
		weapon.Attack(attackerId, target, AttackSkill(attackerId, weapon))
	}

	return true
}

func Attack(attackerId, targetId entity.Id) {
	crit := GetCreature(attackerId)
	// XXX: Fixed the first weapon as preferred melee attack.
	if weapon := weaponLookup[crit.Attack1]; weapon != nil {
		weapon.Attack(attackerId, GetPos(targetId), AttackSkill(attackerId, weapon))
	}
}

func AttackSkill(attackerId entity.Id, weapon *Weapon) float64 {
	// TODO: Special rules for a more fancy attack skill system go here. A
	// default skill of +5 gives around 75 % chance of success, seems to work
	// nicely in practice.
	return 5
}

func DefenseSkill(defenderId entity.Id) float64 {
	// TODO: Defense against different attack types (bullets vs melee).

	// TODO: Actual defense ability.
	return 0
}
