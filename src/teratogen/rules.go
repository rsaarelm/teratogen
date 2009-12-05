package teratogen

import "fmt"
import "math"
import "rand"

import . "gamelib"

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

func Log2Modifier(x int) int {
	absMod := int(Round(Log2(math.Fabs(float64(x)) + 2) - 1));
	return Isignum(x) * absMod;
}

// Smaller things are logarithmically harder to hit.
func MinToHit(scaleDiff int) int {
	return Poor - Log2Modifier(scaleDiff);
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

// The scaleDifference is defender scale - attacker scale.
func IsMeleeHit(toHit, defense int, scaleDifference int) (success bool, degree int) {
	// Hitting requires a minimal absolute success based on the scale of
	// the target and defeating the target's defense ability.
	threshold := MinToHit(scaleDifference);
	hitRoll := FudgeDice() + toHit;
	defenseRoll := FudgeDice() + defense;

	degree = hitRoll - defenseRoll;
	success = hitRoll >= threshold && degree > 0;
	return;
}

func (self *World) Attack(attacker Entity, defender Entity) {
	switch e1 := attacker.(type) {
	case *Creature:
		switch e2 := defender.(type) {
		case *Creature:
			doesHit, hitDegree := IsMeleeHit(
				e1.MeleeSkill, e2.MeleeSkill, e2.Scale - e1.Scale);

			if doesHit {
				// XXX: Assuming melee attack.
				woundLevel := e1.MeleeWoundLevelAgainst(e2, hitDegree);

				if woundLevel > 0 {
					e2.Damage(woundLevel);

					if e2.IsKilledByWounds() {
 						fmt.Fprintf(Msg, "%v killed.\n", Capitalize(e2.Name));
						self.DestroyEntity(defender);
					} else {
 						fmt.Fprintf(Msg, "%v %v.\n",
							Capitalize(e2.Name), e2.WoundDescription());
					}
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