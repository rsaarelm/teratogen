package teratogen

import "fmt"
import "rand"

// Game mechanics stuff.

type ResolutionLevel int const (
	Abysmal = -4 + iota;
	Terrible;
	Poor;
	Mediocre;
	Fair;
	Good;
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
 					fmt.Fprintf(Msg, "%v killed. ", e2.Name);
					self.DestroyEntity(defender);
				} else {
 					fmt.Fprintf(Msg, "%v %v. ",
						e2.Name, WoundDescription(e2.Wounds));
				}
			} else {
 				fmt.Fprintf(Msg, "%v missed. ", e2.Name);
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