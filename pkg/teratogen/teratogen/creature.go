package teratogen

import (
	"exp/iterable"
	"fmt"
	"hyades/geom"
//	"hyades/gfx"
	"hyades/num"
	"hyades/txt"
//	"math"
)

func (self *Blob) MaxWounds() int { return num.Imax(1, (self.GetI(PropToughness)+3)*2+1) }

func (self *Blob) WoundDescription() string {
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
		return "bruised"
	case wounds < 7:
		return "battered"
	}
	// Lots of wounds, but still not really Hurt.
	return "mangled"
}

func (self *Blob) IsKilledByWounds() bool { return self.GetI(PropWounds) > self.MaxWounds() }

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

func (self *Blob) Damage(woundLevel int, cause *Blob) {
	world := GetWorld()
	self.Set(PropWounds, self.GetI(PropWounds)+(woundLevel+1)/2)

//	sx, sy := CenterDrawPos(self.GetPos())
//	go ParticleAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), sx, sy,
//		config.TileScale, 2e8, float64(config.TileScale)*20.0,
//		gfx.Red, gfx.Red, int(20.0*math.Log(float64(woundLevel))/math.Log(2.0)))

	if self.IsKilledByWounds() {
		//PlaySound("death")
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
		//PlaySound("hit")

		Msg("%v %v.\n",
			txt.Capitalize(self.Name), self.WoundDescription())
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
	world := GetWorld()

	if world.IsOpen(self.GetPos().Plus(vec)) {
		self.Move(vec)
		return true
	}
	return false
}

func (self *Blob) CanSeeTo(pos geom.Pt2I) bool {
	dist := 0
	// TODO Customizable max sight range
	sightRange := 18
	for o := range iterable.Drop(geom.Line(self.GetPos(), pos), 1).Iter() {
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
