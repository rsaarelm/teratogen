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

const (
	PropStrength   = "strength"
	PropToughness  = "toughness"
	PropMeleeSkill = "meleeSkill"
	PropScale      = "scale"
	PropWounds     = "wounds"
	PropDensity    = "density"
	FlagObstacle   = "isObstacle"
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
func IsEnemyOf(ent *Entity, possibleEnemy *Entity) bool {
	if ent.GetClass() == PlayerEntityClass &&
		possibleEnemy.GetClass() == EnemyEntityClass {
		return true
	}
	if ent.GetClass() == EnemyEntityClass &&
		possibleEnemy.GetClass() == PlayerEntityClass {
		return true
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

func Attack(attacker *Entity, defender *Entity) {
	doesHit, hitDegree := IsMeleeHit(
		attacker.GetI(PropMeleeSkill), defender.GetI(PropMeleeSkill),
		defender.GetI(PropScale)-attacker.GetI(PropScale))

	if doesHit {
		Msg("%v hits. ", txt.Capitalize(attacker.GetName()))
		// XXX: Assuming melee attack.
		woundLevel := attacker.MeleeWoundLevelAgainst(defender, hitDegree)

		if woundLevel > 0 {
			defender.Damage(woundLevel, attacker)
		} else {
			Msg("%v undamaged.\n", txt.Capitalize(defender.GetName()))
		}
	} else {
		Msg("%v missed.\n", txt.Capitalize(attacker.GetName()))
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
		if e == player {
			continue
		}
		if e.GetClass() == GlobeEntityClass {
			// TODO: Different globe effects.
			if player.GetI(PropWounds) > 0 {
				Msg("The globe bursts. You feel better.\n")
				player.Set(PropWounds, player.GetI(PropWounds)-1)
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
func IsMobile(entity *Entity) bool {
	return entity.GetClass() > CreatureEntityClassStartMarker
}

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

func IsCreature(e *Entity) bool {
	switch e.GetClass() {
	case PlayerEntityClass, EnemyEntityClass:
		return true
	}
	return false
}

func IsTakeableItem(e *Entity) bool { return e.Class == ItemEntityClass }

func TakeItem(subject *Entity, item *Entity) {
	item.InsertSelf(subject)
	Msg("%v takes %v.\n", txt.Capitalize(subject.GetName()), item.GetName())
}

func DropItem(subject *Entity, item *Entity) {
	// TODO: Check if the subject is holding the item.
	item.RemoveSelf()
	item.MoveAbs(subject.GetPos())
	Msg("%v drops %v.\n", txt.Capitalize(subject.GetName()), item.GetName())
}

func SmartPlayerPickup() *Entity {
	world := GetWorld()
	player := world.GetPlayer()
	for ent := range world.EntitiesAt(player.GetPos()) {
		if ent == player {
			continue
		}
		// It's inside something instead of on the floor. Probably
		// already carried by the player.
		if ent.GetParent() != nil {
			continue
		}
		if IsTakeableItem(ent) {
			TakeItem(player, ent)
			return ent
		}
	}
	Msg("Nothing to take here.\n")
	return nil
}
