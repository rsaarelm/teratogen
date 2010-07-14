package teratogen

import (
	"fmt"
	"hyades/entity"
	"hyades/geom"
	"math"
	"rand"
)


const CreatureComponent = entity.ComponentFamily("creature")


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
	Statuses    int32
	// Velocity is the creature's movement vector from it's last turn. Use it
	// for dodge bonuses, charge attacks etc.
	Velocity geom.Vec2I
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
	return fmt.Sprintf("%d health", int(self.Health*100))
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

	adjustedMag := magnitude / self.HealthScale()
	self.Health -= adjustedMag
	if self.Health < 0 {
		self.Die(selfId, causerId)
	}
}

func (self *Creature) Weapon1() *Weapon {
	return weaponLookup[self.Attack1]
}

func (self *Creature) Weapon2() *Weapon {
	return weaponLookup[self.Attack2]
}
