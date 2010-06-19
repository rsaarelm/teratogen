package teratogen

import (
	"fmt"
	"hyades/entity"
	"hyades/geom"
	"hyades/num"
	"math"
	"rand"
)


const CreatureComponent = entity.ComponentFamily("creature")


// Creature intrinsic traits.
const (
	NoIntrinsic   = 0
	IntrinsicSlow = 1 << iota
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
	IntrinsicHorns         // 50 % chance for a bonus +2 piercing melee attack
	IntrinsicHooves        // 50 % chance for a bonus +2 blunt melee attack
	IntrinsicShimmer       // 20 % chance to evade incoming attacks
)

// Creature transient status traits.
const (
	NoStatus   = 0
	StatusSlow = 1 << iota
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
	// The type of damage it is.
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
		Health:     1.0,
		Statuses:   0,
	}
	manager.Handler(CreatureComponent).Add(guid, result)
}


// Creature component. Stats etc.
type Creature struct {
	Power, Skill int
	Scale        int
	Intrinsics   int32
	Health       float64
	Statuses     int32
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

func (self *Creature) Heal(amount float64) {
	self.Health = math.Fmin(1.0, self.Health+amount)
}

// IsHurt returns whether the creature has enough wounds to warrant some
// attention.
func (self *Creature) IsHurt() bool {
	return self.Health <= 0.75
}

func (self *Creature) IsSeriouslyHurt() bool {
	return self.Health <= 0.25
}

func (self *Creature) WoundDescription() string {
	return fmt.Sprintf("%d health", int(self.Health*100))
}

func (self *Creature) IsKilledByWounds() bool { return self.Health < 0.0 }

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

// HealthScale is used to adjust damage dealt to the creature. Creatures take
// 1/HealthScale damage to their 0..1 health value for a single absolute unit
// of damage.
func (self *Creature) HealthScale() float64 {
	return math.Sqrt(math.Pow(2, float64(self.MassFactor()+self.Toughness())))
}

func (self *Creature) Wound(selfId entity.Id, woundLevel int, causerId entity.Id) {
	if self.Statuses&StatusDead != 0 {
		return
	}

	// TODO: Replace woundLevel with a damage system designed for the new
	// health system.
	self.Health -= float64(woundLevel) / 2.0 / self.HealthScale()

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
		EMsg("{sub.Thename} is at {sub.is} %v.\n", selfId, causerId, self.WoundDescription())
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

func (self *Creature) Damage(selfId entity.Id, data *DamageData, hitDegree float64, sourcePos geom.Pt2I, causerId entity.Id) {
	// TODO: Armor can remove damage.
	damageAmount := data.BaseMagnitude
	if hitDegree == 1.0 {
		// Critical hit
		damageAmount *= 2
	}

	isDirectionalDamage := !sourcePos.Equals(GetPos(selfId))
	damageDir := geom.Vec2IToDir6(GetPos(selfId).Minus(sourcePos))

	if data.Type == BluntDamage && isDirectionalDamage {
		// Possibility of knockback.
		knockbackAmount := RollKnockback(data.BaseMagnitude+data.KnockbackBonus, self.MassFactor())
		if knockbackAmount > 0 {
			Knockback(selfId, causerId, damageDir, knockbackAmount)
		}
	}

	if damageAmount > 0 {
		self.Wound(selfId, damageAmount, causerId)
	} else {
		EMsg("{sub.Thename} shrug{sub.s} off the damage.\n", selfId, entity.NilId)
	}

	// XXX: Damage amount not accounted in armor damage.
	DamageEquipment(selfId, ArmorEquipSlot)
}
