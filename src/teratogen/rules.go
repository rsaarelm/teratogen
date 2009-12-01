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

func HitsInMelee(attacker *Creature, defender *Creature) (success int) {
	// Determine hit chance, bigger things are easier to hit.
	toHit := attacker.MeleeSkill - attacker.Scale;
	defend := defender.MeleeSkill - defender.Scale;

	success = FudgeOpposed(toHit, defend);
	return;
}

func (self *World) Attack(attacker Entity, defender Entity) {
	switch e1 := attacker.(type) {
	case *Creature:
		switch e2 := defender.(type) {
		case *Creature:
			hit := HitsInMelee(e1, e2);

			if hit > 0 {
				// XXX: Assuming melee attack.

				damageFactor := e1.Strength + e1.Scale + hit + FudgeDice();
				// TODO: Weapon effects to wound Factor.

				armorFactor := e2.Scale + e2.Toughness;
				// TODO: Armor effects to defense actor.

				woundLevel := IntMax(0, damageFactor - armorFactor);
				e2.Wounds += (woundLevel + 1) / 2;

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

func FudgeOpposed(ability, difficulty int) int {
	return (FudgeDice() + ability) - (FudgeDice() + difficulty);
}