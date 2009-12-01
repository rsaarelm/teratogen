package teratogen

import "fmt"
import "math"
import "rand"

import . "fomalhaut"

// Game mechanics stuff.

type ResolutionLevel int const (
	Abysmal = -4 + iota;
	Terrible;
	Poor;
	Mediocre;
	Fair;
	Good;
	Great;
	Superb;
	Legendary;
)

func (self *Creature) MaxWounds() int {
	return IntMax(1, (self.Toughness + 3) * 2 + 1);
}

// TODO: Make take the creature too, requires toughness
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

func LevelDescription(level int) string {
	switch {
	case level < -4: return fmt.Sprintf("abysmal %d", -(level + 3));
	case level == -4: return "abysmal"
	case level == -3: return "terrible";
	case level == -2: return "poor";
	case level == -1: return "mediocre";
	case level == 0: return "fair";
	case level == 1: return "good";
	case level == 2: return "great";
	case level == 3: return "superb";
	case level == 4: return "legendary";
	case level > 4: return fmt.Sprintf("legendary %d", level - 3);
	}
	panic("Switch fallthrough in LevelDescription");
}

func (self *Creature)IsKilledByWounds() bool {
	return self.Wounds > self.MaxWounds();
}

// Return whether an entity considers another entity an enemy.
func (self *World) IsEnemyOf(ent Entity, possibleEnemy Entity) bool {
	switch e1 := ent.(type) {
	case *Creature:
		switch e2 := possibleEnemy.(type) {
		case *Creature:
			if e1.GetClass() == PlayerEntityClass &&
				e2.GetClass() == EnemyEntityClass {
				return true;
			}
			if e1.GetClass() == EnemyEntityClass &&
				e2.GetClass() == PlayerEntityClass {
				return true;
			}
		}
	}
	return false;
}

func (self *World) Attack(attacker Entity, defender Entity) {
	switch e1 := attacker.(type) {
	case *Creature:
		switch e2 := defender.(type) {
		case *Creature:
			// Determine hit chance, bigger things are easier to hit.
			toHit := e1.Offense - e1.Scale;
			defense := e2.Defense - e2.Scale;
//			fmt.Printf("%v needs %v success to hit %v.\n",
//				Capitalize(e1.Name),
//				LevelDescription(defense - toHit),
//				e2.Name);

			hit := FudgeRoll(toHit, defense);

			if hit > 0 {
				// XXX: Assuming melee attack.

				// Fudge the wound factor with a bonus so that
				// you need a clear armor disadvantage before
				// doing damage becomes uncertain.
				const defaultWoundBonus = 2;
				woundFactor := e1.Strength + e1.Scale + defaultWoundBonus;
				// TODO: Weapon effects to wound Factor.

				armorFactor := e2.Scale;
				// TODO: Armor effects to defense actor.

				damage := IntMax(0, FudgeRoll(woundFactor, armorFactor));
//				fmt.Printf("%v needs %v success to damage %v.\n",
//					Capitalize(e1.Name),
//					LevelDescription(armorFactor - woundFactor),
//					e2.Name);
//
				e2.Wounds += int(math.Ceil(float64(damage / 2)));
				if e2.IsKilledByWounds() {
 					fmt.Fprintf(Msg, "%v killed.\n", Capitalize(e2.Name));
					self.DestroyEntity(defender);
				} else {
 					fmt.Fprintf(Msg, "%v %v.\n",
						Capitalize(e2.Name), e2.WoundDescription());
				}
			} else {
 				fmt.Fprintf(Msg, "%v missed.\n", Capitalize(e2.Name));
			}
		}
	}
}

func FudgeDice() (result int) {
	for i := 0; i < 4; i++ {
		result += -1 + rand.Intn(3);
	}
	return;
}

func FudgeRoll(ability, difficulty int) int {
	return FudgeDice() + ability - difficulty;
}