package teratogen

import (
	"exp/iterable"
	"hyades/entity"
	"hyades/geom"
	"hyades/num"
)

func Attack(attackerId, defenderId entity.Id) {
	attCrit, defCrit := GetCreature(attackerId), GetCreature(defenderId)
	defense := 0
	if defCrit.HasIntrinsic(IntrinsicMartialArtist) {
		defense = defCrit.Skill
	}

	// The larger the defender is relative to the attacker, the easier it is to
	// hit.
	hitDegree := ContestRoll(attCrit.Skill - defense + defCrit.Scale - attCrit.Scale)

	if hitDegree >= 0.0 && defCrit.HasIntrinsic(IntrinsicShimmer) && num.OneChanceIn(5) {
		EMsg("{sub.Thename} phase{sub.s} out of {obj.thename's} reach.\n", defenderId, attackerId)
		return
	}

	if hitDegree >= 0.0 {
		EMsg("{sub.Thename} hit{sub.s} {obj.thename}.\n", attackerId, defenderId)
		// XXX: Assuming melee attack.
		defCrit.Damage(defenderId, attCrit.MeleeDamageData(attackerId), 0,
			GetPos(attackerId), attackerId)

		DamageEquipment(attackerId, MeleeEquipSlot)

		if IsAlive(defenderId) && attCrit.HasIntrinsic(IntrinsicHorns) && num.OneChanceIn(2) {
			EMsg("{sub.Thename} headbutt{sub.s} {obj.thename}.\n", attackerId, defenderId)
			defCrit.Damage(defenderId, &DamageData{2 + attCrit.MassFactor(), PiercingDamage, 0},
				0, GetPos(attackerId), attackerId)
		}

		if IsAlive(defenderId) && attCrit.HasIntrinsic(IntrinsicHooves) && num.OneChanceIn(2) {
			EMsg("{sub.Thename} kick{sub.s} {obj.thename}.\n", attackerId, defenderId)
			defCrit.Damage(defenderId, &DamageData{2 + attCrit.MassFactor(), BluntDamage, 0},
				0, GetPos(attackerId), attackerId)
		}
	} else {
		EMsg("{sub.Thename} {.section sub.you}miss{.or}misses{.end} {obj.thename}.\n",
			attackerId, defenderId)
	}
}

func Knockback(id, causerId entity.Id, dir6 int, amount int) {
	for i := 0; i < amount; i++ {
		if !TryMove(id, geom.Dir6ToVec(dir6)) {
			// Bumped into something, hurt for the amount of movement still left.

			// XXX: If bumped into another creature, should hurt that creature
			// too.
			hurtAmount := amount - i
			if crit := GetCreature(id); crit != nil {
				crit.Damage(id, &DamageData{BaseMagnitude: hurtAmount, Type: BluntDamage},
					0, GetPos(id), causerId)
			}
			return
		}
	}
}

func RollKnockback(attackerPower, defenderMass int) (numCells int) {
	difficulty := defenderMass + 3

	// Attacker needs to keep doing harder and harder strength checks to get
	// more pushback.
	for {
		if ContestRoll(attackerPower-difficulty) >= 0.0 {
			numCells++
			difficulty += 4
		} else {
			break
		}
	}
	return
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

// Shoot makes entity attackerId shoot at target position. Returns whether the
// shooting ends the entity's move.
func Shoot(attackerId entity.Id, target geom.Pt2I) (endsMove bool) {
	if !GunEquipped(attackerId) {
		return true
	}

	// TODO: Aiming precision etc.
	hitPos := GetHitPos(GetPos(attackerId), target)

	damage := &DamageData{Type: PiercingDamage}
	if gun, ok := GetEquipment(attackerId, GunEquipSlot); ok {
		damage.BaseMagnitude += GetItem(gun).WoundBonus
	}

	hitDegree := ContestRoll(GetCreature(attackerId).Skill)

	Fx().Shoot(attackerId, hitPos)

	DamagePos(hitPos, GetPos(attackerId), damage, hitDegree, attackerId)

	DamageEquipment(attackerId, GunEquipSlot)

	if RapidFireGunEquipped(attackerId) {
		endsMove = RapidFireEndsMove()
		if !endsMove {
			EMsg("{sub.Thename} keep{sub.s} shooting.\n", attackerId, entity.NilId)
		}
		return
	}

	return true
}

// RapidFireEndsTurn returns true when a creature shooting with a rapid-fire
// weapon should have it's move ended and false if the creature should be
// allowed to perform another action on its move.
func RapidFireEndsMove() bool { return num.ChancesIn(1, 3) }

func DamageEquipment(ownerId entity.Id, slot EquipSlot) {
	if itemId, ok := GetEquipment(ownerId, slot); ok {
		item := GetItem(itemId)
		if num.OneChanceIn(item.Durability) {
			if slot == GunEquipSlot {
				EMsg("{sub.Thename's} {obj.name} is out of ammo.\n", ownerId, itemId)
			} else {
				EMsg("{sub.Thename's} {obj.name} breaks.\n", ownerId, itemId)
			}
			Destroy(itemId)
		}
	}
}

func DamagePos(pos, sourcePos geom.Pt2I, damage *DamageData, hitDegree float64, causerId entity.Id) {
	for o := range iterable.Filter(EntitiesAt(pos), EntityFilterFn(IsCreature)).Iter() {
		id := o.(entity.Id)
		GetCreature(id).Damage(id, damage, hitDegree, sourcePos, causerId)
	}
}
