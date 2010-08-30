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

const VestArmorLevel = 30
const RiotArmorLevel = 70
const HardsuitArmorLevel = 120

func SpawnWeight(scarcity, minDepth int, depth int) (result float64) {
	const epsilon = 1e-7
	const outOfDepthFactor = 80.0
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

// ContestRoll makes a random opposition check against the given skill rating
// (larger values make easier contests, skill 0 means 50 % probability, a
// change of +/-1 near zero means about +/- 5 % change in probability) and
// returns the degree of success or failure. The result is a number between
// -1.0 and 1.0 inclusive. Positive values are successes, negative values are
// failures, and the absolute value is the degree of success. Values of
// exactly -1.0 or 1.0 can be treated as critical failures and successes.
func ContestRoll(skill float64) (result float64) {
	// Standard deviation of 8 makes a unit of difficulty give around 5 % difference
	const sd = 8.0
	// Critical threshold 1.75 gives around 4 % chances to both critical
	// success and failure.
	const critThres = 1.75

	result = rand.NormFloat64()*sd + skill
	result /= sd * critThres
	result = num.Clamp(-1.0, 1.0, result)
	return
}

// NormRoll gives a normal-distributed integer value around zero which is at
// least -max and at most max. It cuts off the normal distribution at sd=1.75
// and takes a linear sample from the result.
func NormRoll(max int) int {
	const cutoffSd = 1.75

	if max < 0 {
		return 0
	}

	n := max*2 + 1
	x := (rand.NormFloat64() + cutoffSd) / (cutoffSd * 2)
	x = num.Clamp(0.0, 0.999, x)
	return int(x*float64(n)) - max
}

func MovePlayerDir(dir int) {
	GetFov().ClearSight()
	TryMove(PlayerId(), geom.Dir6ToVec(dir))

	GetFov().DoFov(GetPos(PlayerId()))

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
			if GetCreature(PlayerId()).Health < 1.0 {
				EMsg("{obj.Thename} bursts. {sub.Thename} feel{sub.s} better.\n", PlayerId(), id)
				Fx().Heal(PlayerId(), 1)
				GetCreature(PlayerId()).Heal(0.15)
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

// makeSpawnDistribution makes a distribution of the given assemblage set for
// the given depth.
func makeSpawnDistribution(depth int, as iterable.Iterable) num.WeightedDist {
	weightFn := func(item interface{}) float64 {
		metadata := assemblages[item.(string)][Metadata].(*metaTemplate)
		return SpawnWeight(metadata.Scarcity, metadata.MinDepth, depth)
	}
	return num.MakeWeightedDist(weightFn, iterable.Data(as))
}

func BlocksMovement(id entity.Id) bool { return IsCreature(id) }

func Explode(pos geom.Pt2I, power int, cause entity.Id) {
	Fx().Explode(pos, power, 2)
	for pt := range geom.PtIter(pos.X-1, pos.Y-1, 3, 3) {
		DamagePos(pt, pos, float64(power*10), BluntDamage, cause)
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

// ExpValue returns the experience point reward the player should get for
// defeating a creature.
func ExpValue(creatureId entity.Id) float64 {
	if crit := GetCreature(creatureId); crit != nil {
		// XXX: Only looks at hit points, should also be adjusted based on
		// weapons and intrinsics that make the creature more challenging.
		return crit.HealthScale()
	}
	return 0
}

func OnPlayerKill(killedId entity.Id) {
	if mutations := GetMutations(PlayerId()); mutations != nil {
		mutations.GiveExp(PlayerId(), ExpValue(killedId))
	}
}

func IsAlive(id entity.Id) bool {
	if !GetManager().HasEntity(id) {
		return false
	}
	if crit := GetCreature(id); crit != nil {
		if crit.Statuses&StatusDead != 0 {
			return false
		}
	}
	return true
}

func DamagePos(pos, sourcePos geom.Pt2I, magnitude float64, kind DamageType, causerId entity.Id) {
	for o := range iterable.Filter(EntitiesAt(pos), EntityFilterFn(IsCreature)).Iter() {
		id := o.(entity.Id)
		GetCreature(id).Damage(id, causerId, sourcePos, magnitude, kind)
	}
}

// ConfusionScramble randomly rotates a direction a creature is using around
// if the creature has the confused status on. Use this on directions
// creatures use for moving or attacking. Longer than unit length vectors
// retain their length in hexes, but are returned pointing along one of the
// hex cardinal directions if scrambled.
func ConfusionScramble(id entity.Id, dir geom.Vec2I) geom.Vec2I {
	if crit := GetCreature(id); crit != nil && crit.HasStatus(StatusConfused) {
		// Confused creatures don't always manage to step where they want to.
		if num.OneChanceIn(2) {
			// Randomize direction.
			return geom.Dir6ToVec(rand.Intn(6)).Scale(dir.HexLen())
		}
	}
	return dir
}
