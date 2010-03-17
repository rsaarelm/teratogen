package teratogen

import (
	"exp/iterable"
	"fmt"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"hyades/num"
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

func SpawnWeight(scarcity, minDepth int, depth int) (result float64) {
	const epsilon = 1e-7
	const outOfDepthFactor = 2.0
	fscarcity := float64(scarcity)

	if depth < minDepth {
		// Exponentially increase the scarcity for each level out of depth
		outOfDepth := minDepth - depth
		fscarcity *= math.Pow(outOfDepthFactor, float64(outOfDepth))
	}

	const depthCap = 4
	if depth > minDepth+depthCap {
		// Also increase scarcity when we're out of the creature's depth, but
		// only up to a cap.
		outOfDepth := num.Imax(depth-minDepth*2, depthCap)
		fscarcity *= math.Pow(outOfDepthFactor, float64(outOfDepth))
	}

	result = 1.0 / fscarcity
	// Make too scarce weights just plain zero.
	if result < epsilon {
		result = 0.0
	}
	return
}

// Log2Modifier returns the discrete base-2 logarithm class a number belongs
// to. Used for skill values. 0 for 0, signum(x) * floor(log_2(abs(x))) for
// other numbers.
func Log2Modifier(x int) int {
	absMod := int(num.Round(num.Log2(math.Fabs(float64(x))+2) - 1))
	return num.Isignum(x) * absMod
}

// Smaller things are logarithmically harder to hit.
func MinToHit(scaleDiff int) int { return Poor - Log2Modifier(scaleDiff) }

func LevelDescription(level int) string {
	switch {
	case level < -4:
		return fmt.Sprintf("abysmal -%d", -(level + 4))
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
		return fmt.Sprintf("legendary +%d", level-4)
	}
	panic("Switch fallthrough in LevelDescription")
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
	GetLos().ClearSight()
	TryMove(PlayerId(), geom.Dir6ToVec(dir))

	GetLos().DoLos(GetPos(PlayerId()))

	// TODO: More general collision code, do collisions for AI creatures
	// too.

	// See if the player collided with something fun.
	for o := range EntitiesAt(GetPos(PlayerId())).Iter() {
		id := o.(entity.Id)
		if id == entity.NilId {
			continue
		}
		if id == PlayerId() {
			continue
		}
		// TODO: Replace the kludgy special case recognition with a trigger
		// component that gets run when the entity is stepped on.
		if GetName(id) == "globe" {
			// TODO: Different globe effects.
			if GetCreature(PlayerId()).Wounds > 0 {
				mutationMsg := FormatMessage("{obj.Thename} bursts. {sub.Thename} feel{sub.s} strange.\n", PlayerId(), id)
				if !PlayerMutationRoll(1, mutationMsg) {
					EMsg("{obj.Thename} bursts. {sub.Thename} feel{sub.s} better.\n", PlayerId(), id)
				}
				Fx().Heal(PlayerId(), 1)
				GetCreature(PlayerId()).Wounds -= 1
				// Deferring this until the iteration is over.
				defer Destroy(id)
			}
		}
	}
}

func SmartMovePlayer(dir int) {
	// Special 8-directional move, with the straight left/right alternating.
	pos := GetPos(PlayerId())
	column := pos.X - pos.Y
	altDir := -1

	switch dir {
	case 0:
		dir = 0
	case 1:
		dir = 1
	case 2:
		if column%2 == 0 {
			dir = 2
			altDir = 1
		} else {
			dir = 1
			altDir = 2
		}
	case 3:
		dir = 2
	case 4:
		dir = 3
	case 5:
		dir = 4
	case 6:
		if column%2 == 0 {
			dir = 4
			altDir = 5
		} else {
			dir = 5
			altDir = 4
		}
	case 7:
		dir = 5
	}

	target := pos.Plus(geom.Dir6ToVec(dir))

	if IsUnwalkable(target) && altDir != -1 {
		// Alternating gait and the terrain in the target pos isn't good. Go for
		// the alt pos then.
		dir = altDir
		target = pos.Plus(geom.Dir6ToVec(dir))
	}

	for o := range EnemiesAt(PlayerId(), target).Iter() {
		Attack(PlayerId(), o.(entity.Id))
		return
	}
	// No attack, move normally.
	MovePlayerDir(dir)
	StuffOnGroundMsg()
}

// Write a message about interesting stuff on the ground.
func StuffOnGroundMsg() {
	subjectId := PlayerId()
	items := iterable.Data(TakeableItems(GetPos(subjectId)))
	stairs := GetArea().GetTerrain(GetPos(subjectId)) == TerrainStairDown
	if len(items) > 1 {
		Msg("There are several items here.\n")
	} else if len(items) == 1 {
		Msg("There is %v here.\n", GetName(items[0].(entity.Id)))
	}
	if stairs {
		Msg("There are stairs down here.\n")
	}
}

func GameOver(reason string) { Fx().Quit(fmt.Sprintf("You %v\n", reason)) }

func WinGame(message string) { Fx().Quit(fmt.Sprintf("%s\n", message)) }

// Return whether the entity moves around by itself and shouldn't be shown in
// map memory.
func IsMobile(id entity.Id) bool { return IsCreature(id) }

func NumPlayerTakeableItems() int {
	return len(iterable.Data(TakeableItems(GetPos(PlayerId()))))
}

func PlayerAtStairs() bool {
	return GetArea().GetTerrain(GetPos(PlayerId())) == TerrainStairDown
}

func PlayerEnterStairs() {
	if PlayerAtStairs() {
		Msg("Going down...\n")
		NextLevel()
	} else {
		Msg("There are no stairs here.\n")
	}
}

func NextLevel() { GetContext().EnterLevel(GetCurrentLevel() + 1) }

// EntityFilterFn takes a predicate function that works on entity.Ids and
// converts it into a function that works on interface{} values that can be
// used with the iterable API.
func EntityFilterFn(entityPred func(entity.Id) bool) func(interface{}) bool {
	return func(o interface{}) bool { return entityPred(o.(entity.Id)) }
}

// EntityDist returns the distance between two entities.
func EntityDist(id1, id2 entity.Id) float64 {
	if HasPosComp(id1) && HasPosComp(id2) {
		return float64(geom.HexDist(GetPos(id1), GetPos(id2)))
	}
	return math.MaxFloat64
}

const spawnsPerLevel = 32

func Spawn(name string) entity.Id {
	id := assemblages[name].MakeEntity(GetManager())
	return id
}

func SpawnAt(name string, pos geom.Pt2I) (result entity.Id) {
	result = Spawn(name)
	PosComp(result).MoveAbs(pos)
	return
}

func SpawnRandomPos(name string) entity.Id {
	if pos, ok := GetSpawnPos(); ok {
		return SpawnAt(name, pos)
	}
	return entity.NilId
}

func clearNonplayerEntities() {
	// Bring over player object and player's inventory.
	keep := make(map[entity.Id]bool)
	keep[PlayerId()] = true
	for o := range RecursiveContents(PlayerId()).Iter() {
		keep[o.(entity.Id)] = true
	}

	for o := range Entities().Iter() {
		id := o.(entity.Id)
		if _, ok := keep[id]; !ok {
			defer Destroy(id)
		}
	}
}

func makeSpawnDistribution(depth int) num.WeightedDist {
	weightFn := func(item interface{}) float64 {
		metadata := assemblages[item.(string)][Metadata].(*metaTemplate)
		return SpawnWeight(metadata.Scarcity, metadata.MinDepth, depth)
	}
	values := make([]interface{}, len(assemblages))
	i := 0
	for name, _ := range assemblages {
		values[i] = name
		i++
	}
	return num.MakeWeightedDist(weightFn, values)
}

func BlocksMovement(id entity.Id) bool { return IsCreature(id) }

func Explode(pos geom.Pt2I, power int, cause entity.Id) {
	Fx().Explode(pos, power, 2)
	for pt := range geom.PtIter(pos.X-1, pos.Y-1, 3, 3) {
		DamagePos(pt, pos, &DamageData{BaseMagnitude: power, Type: BluntDamage}, 0, cause)
	}
}

// ScaleToVolume converts a scale value into a corresponding volume in liters.
// This is the main interpretation for scale.
func ScaleToVolume(scale float64) (liters float64) {
	return math.Pow(2.0, scale+6)
}

// ScaleToMass converts a scale and a density value (density 0 for water /
// living things) into a corresponding kilogram mass.
func ScaleToMass(scale, density float64) (kg float64) {
	return ScaleToVolume(scale + density)
}

// ScaleToHeight converts a scale to the average upright height of a humanoid
// of that scale in meters.
func ScaleToHeight(scale float64) (meters float64) {
	return math.Pow(math.Pow(2.0, scale+2), 1.0/3.0)
}

type BloodSplatter int

const (
	NoBlood = BloodSplatter(iota)
	BloodTrail
	SmallBloodSplatter
	LargeBloodSplatter
)

func BloodSplatterAt(pos geom.Pt2I) BloodSplatter {
	for o := range EntitiesAt(pos).Iter() {
		id := o.(entity.Id)
		switch GetName(id) {
		case "bloody trail":
			return BloodTrail
		case "blood splatter":
			return SmallBloodSplatter
		case "blood pool":
			return LargeBloodSplatter
		}
	}
	return NoBlood
}

func ClearBloodAt(pos geom.Pt2I) {
	for o := range EntitiesAt(pos).Iter() {
		id := o.(entity.Id)
		switch GetName(id) {
		case "bloody trail", "blood splatter", "blood pool":
			defer Destroy(id)
		}
	}
}

func SplatterBlood(pos geom.Pt2I, amount BloodSplatter) {
	existing := BloodSplatterAt(pos)
	// If it's already bloodier than we'll want to make it.
	if int(existing) >= int(amount) {
		return
	}

	ClearBloodAt(pos)

	var id string
	switch amount {
	case BloodTrail:
		id = "blood_trail"
	case SmallBloodSplatter:
		id = "blood_small"
	case LargeBloodSplatter:
		id = "blood_large"
	default:
		dbg.Warn("Unknown blood spec %v", amount)
	}

	SpawnAt(id, pos)
}

func PlayerIsEsper() bool { return GetCreature(PlayerId()).HasIntrinsic(IntrinsicEsper) }

func CanEsperSense(id entity.Id) bool {
	if crit := GetCreature(id); crit != nil {
		return !crit.HasIntrinsic(IntrinsicUnliving)
	}
	return false
}

func CreaturePowerLevel(id entity.Id) int {
	if crit := GetCreature(id); crit != nil {
		// TODO: Increase power from intrinsincs.
		return num.Imax(1, crit.Power+crit.Scale)
	}
	return 0
}

func OnPlayerKill(killedId entity.Id) {
	const mutationResistance = 10

	power := CreaturePowerLevel(killedId)
	dist := EntityDist(killedId, PlayerId())
	if dist < 2.0 {
		// More effect from close combat kills.
		power *= 2
	}

	PlayerMutationRoll(power, "The kill affects you.\n")
}

func PlayerMutationRoll(power int, msg string) bool {
	mutationResistance := 4 + GetCreature(PlayerId()).Mutations/2

	// Cap too big mutation chances. Always have a reasonably low chance of
	// doing a mutation.
	power = num.Imin(power, mutationResistance-2)

	if FudgeOpposed(power, mutationResistance) >= 0 {
		Msg(msg)
		Fx().MorePrompt()
		Mutate(PlayerId())
		return true
	}
	return false
}
