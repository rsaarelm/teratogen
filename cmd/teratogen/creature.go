package main

import (
	"fmt"
	"hyades/geom"
	"hyades/gfx"
	"hyades/num"
	"hyades/txt"
	"math"
)

func (self *Entity) MaxWounds() int { return num.IntMax(1, (self.GetI(PropToughness)+3)*2+1) }

func (self *Entity) WoundDescription() string {
	maxWounds := self.MaxWounds()
	wounds := self.GetI(PropWounds)
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

func (self *Entity) IsKilledByWounds() bool { return self.GetI(PropWounds) > self.MaxWounds() }

func (self *Entity) MeleeDamageFactor() (result int) {
	result = self.GetI(PropStrength) + self.GetI(PropScale) + self.GetI(PropDensity)
	if o, ok := self.GetGuidOpt(PropMeleeWeaponGuid); ok {
		// Melee weapon bonus
		result += o.GetI(PropWoundBonus)
	}
	return
}

func (self *Entity) ArmorFactor() (result int) {
	result = self.GetI(PropScale) + self.GetI(PropDensity) + self.GetI(PropToughness)
	if o, ok := self.GetGuidOpt(PropBodyArmorGuid); ok {
		// Body armor bonus.
		result += o.GetI(PropDefenseBonus)
	}
	return
}

func (self *Entity) Damage(woundLevel int, cause *Entity) {
	world := GetWorld()
	self.Set(PropWounds, self.GetI(PropWounds)+(woundLevel+1)/2)

	sx, sy := CenterDrawPos(self.GetPos())
	go ParticleAnim(ui.context, ui.AddMapAnim(gfx.NewAnim(0.0)), sx, sy,
		config.TileScale, 2e8, float64(config.TileScale)*20.0,
		gfx.Red, gfx.Red, int(math.Pow(1.5, float64(woundLevel+3))))

	if self.IsKilledByWounds() {
		PlaySound("death")
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
		PlaySound("hit")

		Msg("%v %v.\n",
			txt.Capitalize(self.Name), self.WoundDescription())
	}
}

func (self *Entity) MeleeWoundLevelAgainst(target *Entity, hitDegree int) (woundLevel int) {

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

func (self *Entity) TryMove(vec geom.Vec2I) (success bool) {
	world := GetWorld()

	if world.IsOpen(self.GetPos().Plus(vec)) {
		self.Move(vec)
		return true
	}
	return false
}
