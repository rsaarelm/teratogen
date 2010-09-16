package game

import (
	"fmt"
	"hyades/entity"
	"hyades/geom"
	"hyades/num"
	"math"
	"rand"
)


const CreatureComponent = entity.ComponentFamily("creature")

const ArmorScale = 100

// Creature intrinsic traits.
const (
	NoIntrinsic   = 0
	IntrinsicSlow = 1 << iota
	IntrinsicFast
	IntrinsicDeathsplode
	IntrinsicEndboss
	IntrinsicUnliving      // Creature is not a living thing.
	IntrinsicEsper         // Creature can sense unseen living things.
	IntrinsicMartialArtist // Creature can use it's skill to dodge attacks.
	IntrinsicChaosSpawn    // Creature is a thing of chaos that can't be mutated any further.
	IntrinsicHorns         // 50 % chance for a bonus +2 piercing melee attack
	IntrinsicHooves        // 50 % chance for a bonus +2 blunt melee attack
	IntrinsicShimmer       // 20 % chance to evade incoming attacks
	IntrinsicImmobile
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
	StatusFrozen
)


type CreatureTemplate struct {
	Hp float64
	// Primary attack, generally a melee attack.
	Attack1 int
	// Secondary attack, generally a ranged attack or nothing
	Attack2    int
	Intrinsics int32
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

func (self *CreatureTemplate) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return c
}

func (self *CreatureTemplate) MakeComponent(manager *entity.Manager, guid entity.Id) {
	result := &Creature{
		Attack1:     self.Attack1,
		Attack2:     self.Attack2,
		healthScale: self.Hp,
		Intrinsics:  self.Intrinsics,
		Health:      1.0,
		Armor:       0.0,
		Statuses:    0,
	}
	manager.Handler(CreatureComponent).Add(guid, result)
}


// Creature component. Stats etc.
type Creature struct {
	Attack1     int
	Attack2     int
	Intrinsics  int32
	healthScale float64
	Health      float64
	Armor       float64
	Statuses    int32
	// Velocity is the creature's movement vector from it's last turn. Use it
	// for dodge bonuses, charge attacks etc.
	Velocity  geom.Vec2I
	Cooldowns [NumPowerSlots]int
	Powers    [NumPowerSlots]PowerId
}

func GetCreature(id entity.Id) *Creature {
	if id := GetManager().Handler(CreatureComponent).Get(id); id != nil {
		return id.(*Creature)
	}
	return nil
}

func IsCreature(id entity.Id) bool { return GetCreature(id) != nil }

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
	return fmt.Sprintf("%d health", int(self.Health*self.healthScale))
}

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

// HealthScale is used to adjust damage dealt to the creature. Creatures take
// 1/HealthScale damage to their 0..1 health value for a single absolute unit
// of damage.
func (self *Creature) HealthScale() float64 {
	return self.healthScale
}

func (self *Creature) Die(selfId entity.Id, causerId entity.Id) {
	if self.Statuses&StatusDead != 0 {
		return
	}

	// Mark the critter as dead so whatever happens during it's death
	// doesn't cause a new call to Damage.
	self.Statuses |= StatusDead

	Fx().Destroy(selfId)
	EMsg("{sub.Thename} {sub.is} killed.\n", selfId, causerId)

	// Splatter blood.
	if !self.HasIntrinsic(IntrinsicUnliving) {
		bloodNum := rand.Intn(4)
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
		Explode(GetPos(selfId), 3, selfId)
	}

	Destroy(selfId)
}

func (self *Creature) Damage(selfId, causerId entity.Id, sourcePos geom.Pt2I, magnitude float64, kind DamageType) {
	if self.Armor > 0 {
		// Damage armor if there is one.
		armorMag := magnitude / ArmorScale
		self.Armor -= armorMag

		if self.Armor < 0 {
			// The armor got entirely destroyed.

			// Apply half of the remaining damage to actual health.
			magnitude = -self.Armor * ArmorScale
			magnitude /= 2
			self.Armor = 0
		} else {
			return
		}
	}

	adjustedMag := magnitude / self.HealthScale()

	// Frozen creatures are more death-prone
	if self.Health < adjustedMag*2 && self.HasStatus(StatusFrozen) {
		adjustedMag *= 2
		EMsg("{sub.Thename} shatters!\n", selfId, entity.NilId)
	}

	self.Health -= adjustedMag
	Fx().Damage(selfId, int(math.Log(float64(adjustedMag))/math.Log(2)))
	if self.Health < 0 {
		self.Die(selfId, causerId)
	}
}

// AddArmor checks if the armorPoints provided would increase the creature's
// armor. If so, it sets the armor to the level and returns true. Otherwise
// returns false.
func (self *Creature) AddArmor(armorPoints int) bool {
	level := float64(armorPoints) / ArmorScale
	if level > self.Armor {
		self.Armor = level
		return true
	}
	return false
}

func (self *Creature) Weapon1() *Weapon {
	return weaponLookup[self.Attack1]
}

func (self *Creature) Weapon2() *Weapon {
	return weaponLookup[self.Attack2]
}

// SaveToLose removes the specified status effect with the given probabilty.
// Returns true if the effect was removed, false if the effect stays or wasn't
// present to begin with.
func (self *Creature) SaveToLose(status int32, prob float64) bool {
	if self.HasStatus(status) && rand.Float64() <= prob {
		self.RemoveStatus(status)
		return true
	}
	return false
}

// Heartbeat runs state updates on the creature that happen every turn
// regardless of what the creature is otherwise doing.
func (self *Creature) Heartbeat(selfId entity.Id) {
	self.bloodTrailHeartbeat(selfId)
	self.buffHeartbeat(selfId)

	for i, _ := range self.Cooldowns {
		if self.Cooldowns[i] > 0 {
			self.Cooldowns[i]--
		}
	}
}

func (self *Creature) bloodTrailHeartbeat(selfId entity.Id) {
	standingIn := BloodSplatterAt(GetPos(selfId))
	if standingIn == LargeBloodSplatter {
		// Creatures start tracking blood when they walk through pools of blood.
		self.AddStatus(StatusBloodTrail)
	} else {
		if self.HasStatus(StatusBloodTrail) {
			SplatterBlood(GetPos(selfId), BloodTrail)
			if num.OneChanceIn(3) {
				self.RemoveStatus(StatusBloodTrail)
			}
		}
	}
}

func (self *Creature) buffHeartbeat(selfId entity.Id) {
	if self.SaveToLose(StatusConfused, 1.0/10) {
		EMsg("{sub.Thename} {sub.is} no longer %s.\n",
			selfId, entity.NilId, StatusDescription(StatusConfused))
	}

	if self.SaveToLose(StatusFrozen, 1.0/10) {
		EMsg("{sub.Thename} {sub.is} no longer %s.\n",
			selfId, entity.NilId, StatusDescription(StatusFrozen))
	}
}

func StatusDescription(status int32) string {
	switch status {
	case StatusSlow:
		return "slow"
	case StatusQuick:
		return "quick"
	case StatusConfused:
		return "confused"
	case StatusStunned:
		return "stunned"
	case StatusPoisoned:
		return "poisoned"
	case StatusDead:
		return "dead"
	case StatusBloodTrail:
		return "trailing blood"
	case StatusMutationShield:
		return "protected from mutation"
	case StatusFrozen:
		return "frozen"
	}
	return ""
}
