package main

import (
	"fmt"
	"hyades/geom"
	"hyades/num"
	"hyades/txt"
	"math"
	"rand"
)

// Game mechanics stuff.

type ResolutionLevel int

const (
	Abysmal = -4 + iota
	Terrible
	Poor
	Mediocre
	Fair
	Good
	Great
	Superb
	Legendary
)

func Log2Modifier(x int) int {
	absMod := int(num.Round(num.Log2(math.Fabs(float64(x))+2) - 1))
	return num.Isignum(x) * absMod
}

// Smaller things are logarithmically harder to hit.
func MinToHit(scaleDiff int) int { return Poor - Log2Modifier(scaleDiff) }

func LevelDescription(level int) string {
	switch {
	case level < -4:
		return fmt.Sprintf("abysmal %d", -(level + 3))
	case level == -4:
		return "abysmal"
	case level == -3:
		return "terrible"
	case level == -2:
		return "poor"
	case level == -1:
		return "mediocre"
	case level == 0:
		return "fair"
	case level == 1:
		return "good"
	case level == 2:
		return "great"
	case level == 3:
		return "superb"
	case level == 4:
		return "legendary"
	case level > 4:
		return fmt.Sprintf("legendary %d", level-3)
	}
	panic("Switch fallthrough in LevelDescription")
}

// Return whether an entity considers another entity an enemy.
func IsEnemyOf(ent Entity, possibleEnemy Entity) bool {
	switch e1 := ent.(type) {
	case *Creature:
		switch e2 := possibleEnemy.(type) {
		case *Creature:
			if e1.GetClass() == PlayerEntityClass &&
				e2.GetClass() == EnemyEntityClass {
				return true
			}
			if e1.GetClass() == EnemyEntityClass &&
				e2.GetClass() == PlayerEntityClass {
				return true
			}
		}
	}
	return false
}

// The scaleDifference is defender scale - attacker scale.
func IsMeleeHit(toHit, defense int, scaleDifference int) (success bool, degree int) {
	// Hitting requires a minimal absolute success based on the scale of
	// the target and defeating the target's defense ability.
	threshold := MinToHit(scaleDifference)
	hitRoll := FudgeDice() + toHit
	defenseRoll := FudgeDice() + defense

	degree = hitRoll - defenseRoll
	success = hitRoll >= threshold && degree > 0
	return
}

func Attack(attacker Entity, defender Entity) {
	switch e1 := attacker.(type) {
	case *Creature:
		switch e2 := defender.(type) {
		case *Creature:
			doesHit, hitDegree := IsMeleeHit(
				e1.MeleeSkill, e2.MeleeSkill, e2.Scale-e1.Scale)

			if doesHit {
				Msg("%v hits. ", txt.Capitalize(attacker.GetName()))
				// XXX: Assuming melee attack.
				woundLevel := e1.MeleeWoundLevelAgainst(e2, hitDegree)

				if woundLevel > 0 {
					e2.Damage(woundLevel, e1)
				} else {
					Msg("%v undamaged.", txt.Capitalize(defender.GetName()))
				}
			} else {
				Msg("%v missed.\n", txt.Capitalize(attacker.GetName()))
			}
		}
	}
}

func FudgeDice() (result int) {
	for i := 0; i < 4; i++ {
		result += -1 + rand.Intn(3)
	}
	return
}

func FudgeOpposed(ability, difficulty int) int {
	return (FudgeDice() + ability) - (FudgeDice() + difficulty)
}

func MovePlayerDir(dir int) {
	world := GetWorld()
	player := world.GetPlayer()
	world.ClearLosSight()
	player.TryMove(geom.Dir8ToVec(dir))

	// TODO: More general collision code, do collisions for AI creatures
	// too.

	// See if the player collided with something fun.
	for e := range world.EntitiesAt(player.GetPos()) {
		if e == Entity(player) {
			continue
		}
		if e.GetClass() == GlobeEntityClass {
			// TODO: Different globe effects.
			if player.Wounds > 0 {
				Msg("The globe bursts. You feel better.\n")
				player.Wounds -= 1
				// Deferring this until the iteration is over.
				defer world.DestroyEntity(e)
			}
		}
	}

	world.DoLos(player.GetPos())
}

func SmartMovePlayer(dir int) {
	world := GetWorld()
	player := world.GetPlayer()
	vec := geom.Dir8ToVec(dir)
	target := player.GetPos().Plus(vec)

	for ent := range world.EntitiesAt(target) {
		if IsEnemyOf(player, ent) {
			Attack(player, ent)
			return
		}
	}
	// No attack, move normally.
	MovePlayerDir(dir)
}

func RunAI() {
	world := GetWorld()
	enemyCount := 0
	for crit := range world.IterCreatures() {
		if crit != world.GetPlayer() {
			enemyCount++
		}
		DoAI(crit)
	}
}

func GameOver(reason string) {
	MsgMore()
	fmt.Printf("%v %v\n", txt.Capitalize(GetWorld().GetPlayer().Name), reason)
	Quit()
}

// Return whether the entity moves around by itself and shouldn't be shown in
// map memory.
func IsMobile(entity Entity) bool { return entity.GetClass() > CreatureEntityClassStartMarker }

func PlayerEnterStairs() {
	world := GetWorld()
	if world.GetTerrain(world.GetPlayer().GetPos()) == TerrainStairDown {
		Msg("Going down...\n")
		NextLevel()
	} else {
		Msg("There are no stairs here.\n")
	}
}

func NextLevel() { world.InitLevel(world.CurrentLevelNum() + 1) }
