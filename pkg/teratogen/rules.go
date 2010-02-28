package teratogen

import (
	"exp/iterable"
	"fmt"
	"hyades/alg"
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
func IsEnemyOf(id, possibleEnemyId entity.Id) bool {
	if id == possibleEnemyId {
		return false
	}

	if id == entity.NilId || possibleEnemyId == entity.NilId {
		return false
	}

	// XXX: Currently player is the enemy of every other creature. This should
	// be replaced with a more general faction system.
	if IsCreature(id) && IsCreature(possibleEnemyId) &&
		(id == PlayerId() || possibleEnemyId == PlayerId()) {
		return true
	}

	return false
}

// EnemiesAt iterates the enemies of ent at pos.
func EnemiesAt(id entity.Id, pos geom.Pt2I) iterable.Iterable {
	filter := func(o interface{}) bool { return IsEnemyOf(id, o.(entity.Id)) }

	return iterable.Filter(EntitiesAt(pos), filter)
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

func Attack(attackerId, defenderId entity.Id) {
	attCrit, defCrit := GetCreature(attackerId), GetCreature(defenderId)

	doesHit, hitDegree := IsMeleeHit(attCrit.Melee, defCrit.Melee,
		defCrit.Scale-attCrit.Scale)

	if doesHit {
		Msg("%v hits. ", GetCapName(attackerId))
		// XXX: Assuming melee attack.
		woundLevel := attCrit.MeleeWoundLevelAgainst(
			attackerId, defenderId, hitDegree)

		DamageEquipment(attackerId, MeleeEquipSlot)
		DamageEquipment(defenderId, ArmorEquipSlot)

		if woundLevel > 0 {
			defCrit.Damage(defenderId, woundLevel, attackerId)
		} else {
			Msg("%v undamaged.\n", GetCapName(defenderId))
		}
	} else {
		Msg("%v missed.\n", GetCapName(attackerId))
	}
}

func GetHitPos(origin, target geom.Pt2I) (hitPos geom.Pt2I) {
	for o := range iterable.Drop(geom.Line(origin, target), 1).Iter() {
		hitPos = o.(geom.Pt2I)
		if !IsOpen(hitPos) {
			break
		}
	}
	return
}

func Shoot(attackerId entity.Id, target geom.Pt2I) {
	if !GunEquipped(attackerId) {
		return
	}

	// TODO: Aiming precision etc.
	hitPos := GetHitPos(GetPos(attackerId), target)

	damageFactor := 0
	if gun, ok := GetEquipment(attackerId, GunEquipSlot); ok {
		damageFactor += GetItem(gun).WoundBonus
	}

	Fx().Shoot(attackerId, hitPos)

	DamagePos(hitPos, damageFactor, attackerId)

	DamageEquipment(attackerId, GunEquipSlot)
}

func DamageEquipment(ownerId entity.Id, slot EquipSlot) {
	if itemId, ok := GetEquipment(ownerId, slot); ok {
		item := GetItem(itemId)
		if num.OneChanceIn(item.Durability) {
			if slot == GunEquipSlot {
				Msg("The %s's %s is out of ammo.\n", GetName(ownerId), GetName(itemId))
			} else {
				Msg("The %s's %s breaks.\n", GetName(ownerId), GetName(itemId))
			}
			Destroy(itemId)
		}
	}
}

func DamagePos(pos geom.Pt2I, woundLevel int, causerId entity.Id) {
	for o := range iterable.Filter(EntitiesAt(pos), EntityFilterFn(IsCreature)).Iter() {
		id := o.(entity.Id)
		GetCreature(id).Damage(id, woundLevel, causerId)
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
	GetLos().ClearSight()
	TryMove(PlayerId(), geom.Dir8ToVec(dir))

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
		if GetName(id) == "health globe" {
			// TODO: Different globe effects.
			if GetCreature(PlayerId()).Wounds > 0 {
				Msg("The globe bursts. You feel better.\n")
				Fx().Heal(PlayerId(), 1)
				GetCreature(PlayerId()).Wounds -= 1
				// Deferring this until the iteration is over.
				defer Destroy(id)
			}
		}
	}

	GetLos().DoLos(GetPos(PlayerId()))
}

func SmartMovePlayer(dir int) {
	vec := geom.Dir8ToVec(dir)
	target := GetPos(PlayerId()).Plus(vec)

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

func RunAI() {
	enemyCount := 0
	for o := range Creatures().Iter() {
		id := o.(entity.Id)
		if id != PlayerId() {
			enemyCount++
		}
		DoAI(id)
	}
}

func GameOver(reason string) {
	Fx().Quit(fmt.Sprintf("%v %v\n", GetCapName(PlayerId()), reason))
}

func WinGame(message string) { Fx().Quit(fmt.Sprintf("%s\n", message)) }

// Return whether the entity moves around by itself and shouldn't be shown in
// map memory.
func IsMobile(id entity.Id) bool { return IsCreature(id) }

func PlayerEnterStairs() {
	if GetArea().GetTerrain(GetPos(PlayerId())) == TerrainStairDown {
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

func IsTakeableItem(e entity.Id) bool { return IsItem(e) }

func TakeItem(takerId, itemId entity.Id) {
	SetParent(itemId, takerId)
	Msg("%v takes %v.\n", GetCapName(takerId), GetName(itemId))
}

func DropItem(dropperId, itemId entity.Id) {
	SetParent(itemId, entity.NilId)
	Msg("%v drops %v.\n", GetCapName(dropperId), GetName(itemId))
}

func TakeableItems(pos geom.Pt2I) iterable.Iterable {
	return iterable.Filter(EntitiesAt(pos), EntityFilterFn(IsTakeableItem))
}

func IsEquippableItem(id entity.Id) bool {
	item := GetItem(id)
	return item != nil && item.EquipmentSlot != NoEquipSlot
}

func IsCarryingGear(id entity.Id) bool {
	return iterable.Any(Contents(id), EntityFilterFn(IsEquippableItem))
}

func IsUsable(id entity.Id) bool { return IsItem(id) && GetItem(id).Use != NoUse }

func HasUsableItems(id entity.Id) bool {
	return iterable.Any(Contents(id), EntityFilterFn(IsUsable))
}

func UseItem(userId, itemId entity.Id) {
	if item := GetItem(itemId); item != nil {
		switch item.Use {
		case NoUse:
			Msg("Nothing happens.\n")
		case MedkitUse:
			crit := GetCreature(userId)
			if crit.Wounds > 0 {
				Msg("You feel much better.\n")
				Fx().Heal(userId, crit.Wounds)
				crit.Wounds = 0
				Destroy(itemId)
			} else {
				Msg("You feel fine already.\n")
			}
		default:
			dbg.Die("Unknown use %v.", item.Use)
		}
	}
}

// Autoequip equips item on owner if it can be equpped in a slot that
// currently has nothing.
func AutoEquip(ownerId, itemId entity.Id) {
	slot := GetItem(itemId).EquipmentSlot
	if slot == NoEquipSlot {
		return
	}
	if _, ok := GetEquipment(ownerId, slot); ok {
		// Already got something equipped.
		return
	}
	SetEquipment(ownerId, slot, itemId)
	Msg("Equipped %v.\n", GetName(itemId))
}

func EntityDist(id1, id2 entity.Id) float64 {
	if HasPosComp(id1) && HasPosComp(id2) {
		return GetPos(id1).Minus(GetPos(id2)).Abs()
	}
	return math.MaxFloat64
}

func CreaturesSeenBy(o interface{}) iterable.Iterable {
	id := o.(entity.Id)
	pred := func(o interface{}) bool { return CanSeeTo(GetPos(id), GetPos(o.(entity.Id))) }
	return iterable.Filter(OtherCreatures(o), pred)
}

func ClosestCreatureSeenBy(id entity.Id) entity.Id {
	distFromSelf := func(idOther interface{}) float64 { return EntityDist(idOther.(entity.Id), id) }
	ret, ok := alg.IterMin(CreaturesSeenBy(id), distFromSelf)
	if !ok {
		return entity.NilId
	}
	return ret.(entity.Id)
}

func GunEquipped(id entity.Id) bool {
	_, ok := GetEquipment(id, GunEquipSlot)
	return ok
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

func SpawnRandomPos(name string) entity.Id { return SpawnAt(name, GetSpawnPos()) }

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
		DamagePos(pt, power, cause)
	}
}
