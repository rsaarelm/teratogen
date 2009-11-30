package teratogen

import "fmt"
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

func WoundDescription(wounds int) string {
	switch {
	case wounds > 6:
		return "near death";
	case wounds > 4:
		return "badly hurt";
	case wounds > 2:
		return "hurt";
	case wounds > 0:
		return "grazed";
	}
	return "unhurt";
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

func IsDeadlyWound(wounds int) bool {
	return wounds > 7;
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
			woundFactor, defenseFactor := 0, 0;
			// XXX: Assuming melee attack.
			woundFactor += e1.Strength + e1.Scale;
			// TODO: Weapon effects to wound Factor.
			defenseFactor += e2.Constitution + e2.Scale;
			// TODO: Armor effects to defense factor.

			result := FudgeRoll(woundFactor, defenseFactor);

			if result > 0 {
				e2.Wounds += result;
				if IsDeadlyWound(e2.Wounds) {
 					fmt.Fprintf(Msg, "%v killed.\n", Capitalize(e2.Name));
					self.DestroyEntity(defender);
				} else {
 					fmt.Fprintf(Msg, "%v %v.\n",
						Capitalize(e2.Name), WoundDescription(e2.Wounds));
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