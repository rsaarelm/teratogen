package teratogen

import (
	"hyades/entity"
	"hyades/geom"
	"hyades/num"
	"rand"
)


const CreatureComponent = entity.ComponentFamily("creature")


// Creature intrinsic traits.
const (
	NoIntrinsic = 1 << iota
	IntrinsicSlow
	IntrinsicFast
	IntrinsicBile
	IntrinsicDeathsplode
	IntrinsicPsychicBlast
	IntrinsicConfuse
	IntrinsicElectrocute
	IntrinsicPoison
	IntrinsicEndboss
	IntrinsicTough         // Creature is +2 tougher than it's power.
	IntrinsicFragile       // Creature's toughness is -2 from it's power.
	IntrinsicDense         // Creature's mass is for scale +2 of creature's scale.
	IntrinsicUnliving      // Creature is not a living thing.
	IntrinsicEsper         // Creature can sense unseen living things.
	IntrinsicMartialArtist // Creature can use it's skill to dodge attacks.
	IntrinsicChaosSpawn    // Creature is a thing of chaos that can't be mutated any further.
)

// Creature transient status traits.
const (
	NoStatus = 1 << iota
	StatusSlow
	StatusQuick
	StatusConfused
	StatusStunned
	StatusPoisoned
	StatusDead
	StatusBloodTrail
	StatusMutationShield // Prevents next mutation to creature.
)


type CreatureTemplate struct {
	Power, Skill int
	Scale        int
	Intrinsics   int32
}

type DamageType int

const (
	// Blunt damage does knockback
	BluntDamage = DamageType(iota)
	// Piercing damage doesn't knockback, but can do criticals.
	PiercingDamage

	ElectricDamage
	FireDamage
	ColdDamage
	AcidDamage
)

type DamageData struct {
	// The base magnitude of damage.
	BaseMagnitude int
	// How skillfully the damage was targeted.
	Type DamageType
	// Is there extra knockback involved?
	KnockbackBonus int
}

func (self *CreatureTemplate) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return c
}

func (self *CreatureTemplate) MakeComponent(manager *entity.Manager, guid entity.Id) {
	result := &Creature{
		Power:      self.Power,
		Skill:      self.Skill,
		Scale:      self.Scale,
		Intrinsics: self.Intrinsics,
		Wounds:     0,
		Statuses:   0,
		Mutations:  0,
	}
	manager.Handler(CreatureComponent).Add(guid, result)
}


// Creature component. Stats etc.
type Creature struct {
	Power, Skill int
	Scale        int
	Intrinsics   int32
	Wounds       int
	Statuses     int32
	Mutations    int
}

func GetCreature(id entity.Id) *Creature {
	if id := GetManager().Handler(CreatureComponent).Get(id); id != nil {
		return id.(*Creature)
	}
	return nil
}

func IsCreature(id entity.Id) bool { return GetCreature(id) != nil }

func (self *Creature) Toughness() (result int) {
	result = self.Power
	if self.HasIntrinsic(IntrinsicTough) {
		result += 2
	}
	if self.HasIntrinsic(IntrinsicFragile) {
		result -= 2
	}
	return
}

func (self *Creature) MaxWounds() int { return num.Imax(1, (self.Toughness()+3)*2+1) }

// IsHurt returns whether the creature has enough wounds to warrant some
// attention.
func (self *Creature) IsHurt() bool {
	return self.Wounds > 0 && self.MaxWounds()-self.Wounds < 6
}

func (self *Creature) IsSeriouslyHurt() bool {
	return self.Wounds > 0 && self.MaxWounds()-self.Wounds < 3
}

func (self *Creature) WoundDescription() string {
	maxWounds := self.MaxWounds()
	switch {
	// Statuses where the creature is seriously hurt.
	case maxWounds-self.Wounds < 2:
		return "near death"
	case maxWounds-self.Wounds < 4:
		return "badly hurt"
	case maxWounds-self.Wounds < 6:
		return "hurt"
	// Now describing grazed statuses, which there can be more if the
	// creature is very tough and takes a long time to get to Hurt.
	case self.Wounds < 1:
		return "unhurt"
	case self.Wounds < 3:
		return "grazed"
	case self.Wounds < 5:
		return "bruised"
	case self.Wounds < 7:
		return "battered"
	}
	// Lots of wounds, but still not really Hurt.
	return "mangled"
}

func (self *Creature) IsKilledByWounds() bool { return self.Wounds > self.MaxWounds() }

func (self *Creature) AddIntrinsic(intrinsic int32) {
	self.Intrinsics |= intrinsic
}

func (self *Creature) HasIntrinsic(intrinsic int32) bool {
	return self.Intrinsics&intrinsic != 0
}

func (self *Creature) HasStatus(status int32) bool {
	return self.Statuses&status != 0
}

func (self *Creature) AddStatus(status int32) { self.Statuses |= status }

func (self *Creature) RemoveStatus(status int32) {
	self.Statuses &^= status
}

func (self *Creature) MeleeDamageFactor(id entity.Id) (result int) {
	result = self.Power + self.MassFactor()
	if o, ok := GetEquipment(id, MeleeEquipSlot); ok {
		// Melee weapon bonus
		result += GetItem(o).WoundBonus
	}
	return
}

func (self *Creature) MassFactor() (mass int) {
	mass = self.Scale
	if self.HasIntrinsic(IntrinsicDense) {
		mass += 2
	}
	return
}

func (self *Creature) ArmorFactor(id entity.Id) (result int) {
	result = self.MassFactor() + self.Toughness()
	if o, ok := GetEquipment(id, ArmorEquipSlot); ok {
		// Body armor bonus.
		result += GetItem(o).DefenseBonus
	}
	return
}

func (self *Creature) Wound(selfId entity.Id, woundLevel int, causerId entity.Id) {
	if self.Statuses&StatusDead != 0 {
		return
	}

	woundAmount := (woundLevel + 1) / 2
	self.Wounds += woundAmount

	if self.IsKilledByWounds() {
		// Mark the critter as dead so whatever happens during it's death
		// doesn't cause a new call to Damage.
		self.Statuses |= StatusDead

		Fx().Destroy(selfId)
		EMsg("{sub.Thename} {sub.is} killed.\n", selfId, causerId)

		// Splatter blood.
		if !self.HasIntrinsic(IntrinsicUnliving) {
			bloodNum := rand.Intn(3 + self.Scale)
			if bloodNum >= 2 {
				SplatterBlood(GetPos(selfId), LargeBloodSplatter)
			} else {
				SplatterBlood(GetPos(selfId), SmallBloodSplatter)
			}
		}

		if selfId == PlayerId() {
			var msg string
			if causerId != entity.NilId {
				msg = FormatMessage("were killed by {sub.aname}.", causerId, entity.NilId)
			} else {
				msg = "died."
			}
			GameOver(msg)
		}

		if self.Intrinsics&IntrinsicEndboss != 0 {
			// Killing the endboss.
			WinGame("You win the game, hooray.")
		}

		if causerId == PlayerId() {
			OnPlayerKill(selfId)
		}

		// Deathsplosion.
		if self.Intrinsics&IntrinsicDeathsplode != 0 {
			EMsg("{sub.Thename} blow{sub.s} up!\n", selfId, causerId)
			Explode(GetPos(selfId), 3+self.Scale, selfId)
		}

		Destroy(selfId)
	} else {
		Fx().Damage(selfId, woundLevel)
		EMsg("{sub.Thename} {sub.is} %v.\n", selfId, causerId, self.WoundDescription())
	}
}

func (self *Creature) MeleeWoundLevelAgainst(id, targetId entity.Id, hitDegree int) (woundLevel int) {
	damageFactor := self.MeleeDamageFactor(id) + hitDegree

	armorFactor := GetCreature(targetId).ArmorFactor(targetId)

	woundLevel = damageFactor - armorFactor

	// Not doing any wounds even though hit was successful. Mostly this is
	// when a little critter tries to hit a big one.
	if woundLevel < 1 {
		// If you scored a good hit, you get one chance in the amount
		// woundLevel went below 1 to hit anyway.
		if hitDegree > Log2Modifier(-woundLevel) &&
			num.OneChanceIn(1-woundLevel) {
			woundLevel = 1
		} else {
			woundLevel = 0
		}
	}
	return
}

func (self *Creature) MeleeDamageData(selfId entity.Id) (result *DamageData) {
	result = new(DamageData)

	result.BaseMagnitude = self.Power + self.MassFactor()
	result.Type = BluntDamage

	if o, ok := GetEquipment(selfId, MeleeEquipSlot); ok {
		item := GetItem(o)
		// Melee weapon bonus
		result.BaseMagnitude += item.WoundBonus

		if item.HasTrait(ItemKnockback) {
			result.KnockbackBonus += ItemKnockbackStrength
		}
	}

	return
}

// Deal damage to a creature. BaseDamage is the basic strength of the attack,
// hitDegree is the skill value (0 being neutral, larger being better) with
// which the attack was made. A skilled attack has a small chance of causing a
// wound even if the baseDamage would not otherwise harm the creature.
func (self *Creature) Damage(selfId entity.Id, data *DamageData, hitDegree int, sourcePos geom.Pt2I, causerId entity.Id) {
	damageFactor := data.BaseMagnitude + hitDegree
	armorFactor := self.ArmorFactor(selfId)

	woundLevel := damageFactor - armorFactor

	isDirectionalDamage := !sourcePos.Equals(GetPos(selfId))
	damageDir := geom.Vec2IToDir6(GetPos(selfId).Minus(sourcePos))

	if woundLevel < 1 {
		// If wounding by normal means didn't work, there's still a chance to
		// score a wound if hitDegree was exceptionally good. Hit degree must
		// beat the base-2 logarithm of the negative wound level's magnitude.
		magnitude := num.Iabs(woundLevel)
		if hitDegree > Log2Modifier(magnitude) && num.OneChanceIn(1+magnitude) {
			woundLevel = 1
		} else {
			woundLevel = 0
		}
	}

	if data.Type == BluntDamage && isDirectionalDamage {
		// Possibility of knockback.
		knockbackAmount := RollKnockback(data.BaseMagnitude+data.KnockbackBonus, self.MassFactor())
		if knockbackAmount > 0 {
			Knockback(selfId, causerId, damageDir, knockbackAmount)
		}
	}

	if woundLevel > 0 {
		self.Wound(selfId, woundLevel, causerId)
	} else {
		EMsg("{sub.Thename} shrug{sub.s} off the damage.\n", selfId, entity.NilId)
	}

	// XXX: Damage amount not accounted in armor damage.
	DamageEquipment(selfId, ArmorEquipSlot)
}
