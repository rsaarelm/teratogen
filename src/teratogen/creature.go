package teratogen

import "fmt"

import . "gamelib"

type Creature struct {
	*Icon;
	guid Guid;
	Name string;
	pos Pt2I;
	class EntityClass;
	Strength int;
	Scale int;
	// Added to Scale to determine strength modifier and damage
	// resistance.
	Density int;
	Toughness int;
	Wounds int;
	MeleeSkill int;
}

func (self *Creature) IsObstacle() bool { return true }

func (self *Creature) GetPos() Pt2I { return self.pos }

func (self *Creature) GetGuid() Guid { return self.guid }

func (self *Creature) GetClass() EntityClass { return self.class; }

func (self *Creature) GetName() string { return self.Name; }

// XXX: Assuming Pt2I to be a value type here.
func (self *Creature) MoveAbs(pos Pt2I) { self.pos = pos }

func (self *Creature) Move(vec Vec2I) { self.pos = self.pos.Plus(vec) }

func (self *Creature) MaxWounds() int {
	return IntMax(1, (self.Toughness + 3) * 2 + 1);
}

func (self *Creature) WoundDescription() string {
	maxWounds := self.MaxWounds();
	wounds := self.Wounds;
	switch {
	case maxWounds - wounds < 2: return "near death";
	case maxWounds - wounds < 4: return "badly hurt";
	case maxWounds - wounds < 6: return "hurt";
	// Now describing grazed statuses, which there can be more if the
	// creature is very tough and takes a long time to get to Hurt.
	case wounds < 1: return "unhurt"
	case wounds < 3: return "grazed";
	case wounds < 5: return "cut";
	case wounds < 7: return "battered";
	}
	// Lots of wounds, but still not really Hurt.
	return "mangled";
}

func (self *Creature)IsKilledByWounds() bool {
	return self.Wounds > self.MaxWounds();
}

func (self *Creature) MeleeDamageFactor() int {
	return self.Strength + self.Scale + self.Density;
	// TODO: Weapon effect.
}

func (self *Creature) ArmorFactor() int {
	return self.Scale + self.Density + self.Toughness;
	// TODO: Effects from worn armor.
}

func (self *Creature) Damage(woundLevel int, cause Entity) {
	world := GetWorld();
	self.Wounds += (woundLevel + 1) / 2;

	if self.IsKilledByWounds() {
		if self == world.GetPlayer() {
			Msg("You die.\n");
			var msg string;
			if cause != nil {
				msg = fmt.Sprintf("killed by %v.", cause.GetName());
			} else {
				msg = "killed.";
			}
			GameOver(msg);
		} else {
 			Msg("%v killed.\n", Capitalize(self.Name));
		}
		world.DestroyEntity(self);
	} else {
 		Msg("%v %v.\n",
			Capitalize(self.Name), self.WoundDescription());
	}
}

func (self *Creature) MeleeWoundLevelAgainst(
	target *Creature, hitDegree int) (woundLevel int) {

	damageFactor := self.MeleeDamageFactor() + hitDegree;

	armorFactor := target.ArmorFactor();

	woundLevel = damageFactor - armorFactor;

	// Not doing any wounds even though hit was successful. Mostly this is
	// when a little critter tries to hit a big one.
	if woundLevel < 1 {
		// If you scored a good hit, you get one chance in the amount
		// woundLevel went below 1 to hit anyway.
		if hitDegree > Log2Modifier(-woundLevel) &&
			OneChanceIn(1 - woundLevel) {
			woundLevel = 1;
		} else {
			woundLevel = 0;
		}
	}
	return;
}

func (self *Creature) TryMove(vec Vec2I) (success bool) {
	world := GetWorld();

	if world.IsOpen(self.GetPos().Plus(vec)) {
		self.Move(vec);
		return true;
	}
	return false;
}