package main

import (
	"fmt"
	"hyades/geom"
	"hyades/num"
	"hyades/txt"
)

func (self *EntityBase) MaxWounds() int {
	return num.IntMax(1, (self.Get(PropToughness).(int)+3)*2+1)
}

func (self *EntityBase) WoundDescription() string {
	maxWounds := self.MaxWounds()
	wounds := self.Get(PropWounds).(int)
	switch {
	case maxWounds-wounds < 2:
		return "near death"
	case maxWounds-wounds < 4:
		return "badly hurt"
	case maxWounds-wounds < 6:
		return "hurt"
	// Now describing grazed statuses, which there can be more if the
	// creature is very tough and takes a long time to get to Hurt.
	case wounds < 1:
		return "unhurt"
	case wounds < 3:
		return "grazed"
	case wounds < 5:
		return "cut"
	case wounds < 7:
		return "battered"
	}
	// Lots of wounds, but still not really Hurt.
	return "mangled"
}

func (self *EntityBase) IsKilledByWounds() bool {
	return self.Get(PropWounds).(int) > self.MaxWounds()
}

func (self *EntityBase) MeleeDamageFactor() int {
	return self.Get(PropStrength).(int) + self.Get(PropScale).(int) + self.Get(PropDensity).(int)
	// TODO: Weapon effect.
}

func (self *EntityBase) ArmorFactor() int {
	return self.Get(PropScale).(int) + self.Get(PropDensity).(int) + self.Get(PropToughness).(int)
	// TODO: Effects from worn armor.
}

func (self *EntityBase) Damage(woundLevel int, cause Entity) {
	world := GetWorld()
	self.Set(PropWounds, self.Get(PropWounds).(int)+(woundLevel+1)/2)

	if self.IsKilledByWounds() {
		if self == world.GetPlayer() {
			Msg("You die.\n")
			var msg string
			if cause != nil {
				msg = fmt.Sprintf("killed by %v.", cause.GetName())
			} else {
				msg = "killed."
			}
			GameOver(msg)
		} else {
			Msg("%v killed.\n", txt.Capitalize(self.Name))
		}
		world.DestroyEntity(self)
	} else {
		Msg("%v %v.\n",
			txt.Capitalize(self.Name), self.WoundDescription())
	}
}

func (self *EntityBase) MeleeWoundLevelAgainst(target *EntityBase, hitDegree int) (woundLevel int) {

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

func (self *EntityBase) TryMove(vec geom.Vec2I) (success bool) {
	world := GetWorld()

	if world.IsOpen(self.GetPos().Plus(vec)) {
		self.Move(vec)
		return true
	}
	return false
}
