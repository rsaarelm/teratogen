package main

import (
	"exp/iterable"
	"fmt"
	"hyades/entity"
	"hyades/geom"
	"hyades/gfx"
	"hyades/num"
	"math"
)


const CreatureComponent = entity.ComponentFamily("creature")


type CreatureTemplate struct {
	Str, Tough, Melee int
	Scale, Density    int
}

func (self *CreatureTemplate) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return c
}

func (self *CreatureTemplate) MakeComponent(manager *entity.Manager, guid entity.Id) {
	result := &Creature{
		Str: self.Str,
		Tough: self.Tough,
		Melee: self.Melee,
		Scale: self.Scale,
		Density: self.Density,
		Wounds: 0}
	GetManager().Handler(CreatureComponent).Add(guid, result)
}


// Creature component. Stats etc.
type Creature struct {
	Str, Tough, Melee int
	Scale, Density    int
	Wounds            int
}

func GetCreature(id entity.Id) *Creature {
	return GetManager().Handler(CreatureComponent).Get(id).(*Creature)
}

func (self *Creature) MaxWounds() int { return num.Imax(1, (self.Tough+3)*2+1) }

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

func (self *Blob) MeleeDamageFactor() (result int) {
	result = self.GetI(PropStrength) + self.GetI(PropScale) + self.GetI(PropDensity)
	if o, ok := GetEquipment(self.GetGuid(), PropMeleeWeaponGuid); ok {
		// Melee weapon bonus
		result += GetBlobs().Get(o).(*Blob).GetI(PropWoundBonus)
	}
	return
}

func (self *Blob) ArmorFactor() (result int) {
	result = self.GetI(PropScale) + self.GetI(PropDensity) + self.GetI(PropToughness)
	if o, ok := GetEquipment(self.GetGuid(), PropBodyArmorGuid); ok {
		// Body armor bonus.
		result += GetBlobs().Get(o).(*Blob).GetI(PropDefenseBonus)
	}
	return
}

func (self *Blob) Damage(woundLevel int, causerId entity.Id) {
	GetCreature(self.GetGuid()).Wounds += (woundLevel + 1) / 2

	sx, sy := CenterDrawPos(GetPos(self.GetGuid()))
	go ParticleAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), sx, sy,
		config.TileScale, 2e8, float64(config.TileScale)*20.0,
		gfx.Red, gfx.Red, int(20.0*math.Log(float64(woundLevel))/math.Log(2.0)))

	if GetCreature(self.GetGuid()).IsKilledByWounds() {
		PlaySound("death")
		if self.GetGuid() == PlayerId() {
			Msg("You die.\n")
			var msg string
			if causerId != entity.NilId {
				msg = fmt.Sprintf("killed by %v.", GetName(causerId))
			} else {
				msg = "killed."
			}
			GameOver(msg)
		} else {
			Msg("%v killed.\n", GetCapName(self.GetGuid()))
		}
		DestroyBlob(self)
	} else {
		PlaySound("hit")

		Msg("%v %v.\n", GetCapName(self.GetGuid()), GetCreature(self.GetGuid()).WoundDescription())
	}
}

func (self *Blob) MeleeWoundLevelAgainst(target *Blob, hitDegree int) (woundLevel int) {
	damageFactor := self.MeleeDamageFactor() + hitDegree

	armorFactor := target.ArmorFactor()

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

func (self *Blob) TryMove(vec geom.Vec2I) (success bool) {
	if IsOpen(GetPos(self.GetGuid()).Plus(vec)) {
		PosComp(self.GetGuid()).Move(vec)
		return true
	}
	return false
}

func (self *Blob) CanSeeTo(pos geom.Pt2I) bool {
	dist := 0
	// TODO Customizable max sight range
	sightRange := 18
	for o := range iterable.Drop(geom.Line(GetPos(self.GetGuid()), pos), 1).Iter() {
		if dist > sightRange {
			return false
		}
		dist++
		pt := o.(geom.Pt2I)
		// Can see to the final cell even if that cell does block further sight.
		if pt.Equals(pos) {
			break
		}
		if GetArea().BlocksSight(pt) {
			return false
		}
	}
	return true
}
