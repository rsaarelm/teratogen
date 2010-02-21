package main

import (
	"fmt"
	"hyades/entity"
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
	manager.Handler(CreatureComponent).Add(guid, result)
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

func (self *Creature) MeleeDamageFactor(id entity.Id) (result int) {
	result = self.Str + self.Scale + self.Density
	if o, ok := GetEquipment(id, MeleeEquipSlot); ok {
		// Melee weapon bonus
		result += GetItem(o).WoundBonus
	}
	return
}

func (self *Creature) ArmorFactor(id entity.Id) (result int) {
	result = self.Scale + self.Density + self.Tough
	if o, ok := GetEquipment(id, ArmorEquipSlot); ok {
		// Body armor bonus.
		result += GetItem(o).DefenseBonus
	}
	return
}

func (self *Creature) Damage(id entity.Id, woundLevel int, causerId entity.Id) {
	self.Wounds += (woundLevel + 1) / 2

	sx, sy := CenterDrawPos(GetPos(id))
	go ParticleAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), sx, sy,
		config.TileScale, 2e8, float64(config.TileScale)*20.0,
		gfx.Red, gfx.Red, int(20.0*math.Log(float64(woundLevel))/math.Log(2.0)))

	if self.IsKilledByWounds() {
		PlaySound("death")
		if id == PlayerId() {
			Msg("You die.\n")
			var msg string
			if causerId != entity.NilId {
				msg = fmt.Sprintf("killed by %v.", GetName(causerId))
			} else {
				msg = "killed."
			}
			GameOver(msg)
		} else {
			Msg("%v killed.\n", GetCapName(id))
		}
		DestroyBlob(GetBlob(id))
	} else {
		PlaySound("hit")

		Msg("%v %v.\n", GetCapName(id), self.WoundDescription())
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
