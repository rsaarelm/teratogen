package main

import (
	"exp/draw"
	"exp/iterable"
	"fmt"
	"hyades/alg"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"hyades/gfx"
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

const (
	PropItemUse = "itemUse"

	FlagObstacle = "isObstacle"
)

// Item use type
const (
	NoUse = iota
	MedkitUse
)

// Put item classes before creature classes, so we can use this to control
// draw order as well.
type EntityClass int

// XXX: Save compatibility is easily broken with this as adding things in the
// categories displaces the values further up.

const (
	EmptyEntityClass EntityClass = iota

	// Item classes
	GlobeEntityClass // Globe items are used when stepped on.
	ItemEntityClass  // Items can be picked up and dropped.

	// Creature classes
	CreatureEntityClassStartMarker

	PlayerEntityClass
	EnemyEntityClass
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
	ent, possibleEnemy := GetBlob(id), GetBlob(possibleEnemyId)
	if ent == nil || possibleEnemy == nil {
		return false
	}

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

func Shoot(attackerId entity.Id, target geom.Pt2I) {
	attacker := GetBlob(attackerId)
	if !GunEquipped(attacker) {
		return
	}

	// TODO: Aiming precision etc.
	origin := GetPos(attacker.GetGuid())
	var hitPos geom.Pt2I
	for o := range iterable.Drop(geom.Line(origin, target), 1).Iter() {
		hitPos = o.(geom.Pt2I)
		if !IsOpen(hitPos) {
			break
		}
	}

	damageFactor := 0
	if gun, ok := GetEquipment(attacker.GetGuid(), GunEquipSlot); ok {
		damageFactor += GetItem(gun).WoundBonus
	}

	p1, p2 := draw.Pt(Tile2WorldPos(GetPos(PlayerId()))), draw.Pt(Tile2WorldPos(hitPos))
	go LineAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), p1, p2, 2e8, gfx.White, gfx.DarkRed, config.Scale*config.TileScale)

	// TODO: Sparks when hitting walls.
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
	player := GetBlob(PlayerId())
	GetLos().ClearSight()
	TryMove(PlayerId(), geom.Dir8ToVec(dir))

	// TODO: More general collision code, do collisions for AI creatures
	// too.

	// See if the player collided with something fun.
	for o := range EntitiesAt(GetPos(player.GetGuid())).Iter() {
		id := o.(entity.Id)
		ent := GetBlob(id)
		if ent == nil {
			continue
		}
		if ent == player {
			continue
		}
		if ent.GetClass() == GlobeEntityClass {
			// TODO: Different globe effects.
			if GetCreature(PlayerId()).Wounds > 0 {
				Msg("The globe bursts. You feel better.\n")
				PlaySound("heal")
				GetCreature(PlayerId()).Wounds -= 1
				// Deferring this until the iteration is over.
				defer DestroyBlob(ent)
			}
		}
	}

	GetLos().DoLos(GetPos(player.GetGuid()))
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
	MsgMore()
	fmt.Printf("%v %v\n", GetCapName(PlayerId()), reason)
	Quit()
}

// Return whether the entity moves around by itself and shouldn't be shown in
// map memory.
func IsMobile(id entity.Id) bool {
	if entity := GetBlob(id); entity != nil {
		return entity.GetClass() > CreatureEntityClassStartMarker
	}
	return false
}

func PlayerEnterStairs() {
	if GetArea().GetTerrain(GetPos(PlayerId())) == TerrainStairDown {
		Msg("Going down...\n")
		NextLevel()
	} else {
		Msg("There are no stairs here.\n")
	}
}

func NextLevel() { GetContext().EnterLevel(GetCurrentLevel()) }

// EntityFilterFn takes a predicate function that works on entity.Ids and
// converts it into a function that works on interface{} values that can be
// used with the iterable API.
func EntityFilterFn(entityPred func(entity.Id) bool) (func(interface{}) bool) {
	return func(o interface{}) bool { return entityPred(o.(entity.Id)) }
}

func IsCreature(id entity.Id) bool {
	if ent := GetBlob(id); ent != nil {
		switch ent.GetClass() {
		case PlayerEntityClass, EnemyEntityClass:
			return true
		}
	}
	return false
}

func IsTakeableItem(e *Blob) bool { return e.Class == ItemEntityClass }

func TakeItem(subject *Blob, item *Blob) {
	SetParent(item.GetGuid(), subject.GetGuid())
	Msg("%v takes %v.\n", GetCapName(subject.GetGuid()), GetName(item.GetGuid()))
}

func DropItem(subject *Blob, item *Blob) {
	// TODO: Check if the subject is holding the item.
	SetParent(item.GetGuid(), entity.NilId)
	PosComp(item.GetGuid()).MoveAbs(GetPos(subject.GetGuid()))
	Msg("%v drops %v.\n", GetCapName(subject.GetGuid()), GetName(item.GetGuid()))
}

func TakeableItems(pos geom.Pt2I) iterable.Iterable {
	return iterable.Filter(EntitiesAt(pos), func(o interface{}) bool { return IsTakeableItem(GetBlob(o.(entity.Id))) })
}

// TODO: Change other functions to use interface{} instead of *Blob to make
// it easier to use them with iterable functions.

func IsEquippableItem(id entity.Id) bool {
	item := GetItem(id)
	return item != nil && item.EquipmentSlot != NoEquipSlot
}

func IsCarryingGear(o interface{}) bool {
	return iterable.Any(Contents(o.(*Blob).GetGuid()), EntityFilterFn(IsEquippableItem))
}

func IsCarryingGearFor(o interface{}, slot EquipSlot) bool {
	return iterable.Any(Contents(o.(*Blob).GetGuid()), func(item interface{}) bool { return CanEquipIn(slot, item.(entity.Id)) })
}

func IsUsable(id entity.Id) bool { return GetBlob(id).Has(PropItemUse) }

func HasUsableItems(o interface{}) bool {
	return iterable.Any(Contents(o.(*Blob).GetGuid()), EntityFilterFn(IsUsable))
}

func UseItem(user *Blob, item *Blob) {
	if use, ok := item.GetIOpt(PropItemUse); ok {
		switch use {
		case NoUse:
			Msg("Nothing happens.\n")
		case MedkitUse:
			crit := GetCreature(user.GetGuid())
			if crit.Wounds > 0 {
				Msg("You feel much better.\n")
				PlaySound("heal")
				crit.Wounds = 0
				DestroyBlob(item)
			} else {
				Msg("You feel fine already.\n")
			}
		default:
			dbg.Die("Unknown use %v.", use)
		}
	}
}

func SmartPlayerPickup(alwaysPickupFirst bool) *Blob {
	player := GetBlob(PlayerId())
	itemIds := iterable.Data(TakeableItems(GetPos(player.GetGuid())))
	// XXX: Blob kludge
	items := make([]interface{}, len(itemIds))
	for i := 0; i < len(items); i++ {
		items[i] = GetBlob(itemIds[i].(entity.Id))
	}

	// TODO: Drop blob kludge, rewrite to use MultiChoiceDialog, since ids
	// don't name themselves nicey.
	if len(items) == 0 {
		Msg("Nothing to take here.\n")
		return nil
	}

	choice := items[0]
	if len(items) > 1 && !alwaysPickupFirst {
		var ok bool
		choice, ok = ObjectChoiceDialog("Pick up which item?", items)
		if !ok {
			Msg("Okay, then.\n")
			return nil
		}
	}
	ent := choice.(*Blob)
	TakeItem(player, ent)
	AutoEquip(player, ent)
	return ent
}

// Autoequip equips item on owner if it can be equpped in a slot that
// currently has nothing.
func AutoEquip(owner *Blob, item *Blob) {
	slot := GetItem(item.GetGuid()).EquipmentSlot
	if slot == NoEquipSlot {
		return
	}
	if _, ok := GetEquipment(owner.GetGuid(), slot); ok {
		// Already got something equipped.
		return
	}
	SetEquipment(owner.GetGuid(), slot, item.GetGuid())
	Msg("Equipped %v.\n", item)
}

func EntityDist(o1, o2 interface{}) float64 {
	id1, id2 := o1.(entity.Id), o2.(entity.Id)

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

func ClosestCreatureSeenBy(o interface{}) entity.Id {
	distFromSelf := func(o1 interface{}) float64 { return EntityDist(o1, o) }
	ret, ok := alg.IterMin(CreaturesSeenBy(o), distFromSelf)
	if !ok {
		return entity.NilId
	}
	return ret.(entity.Id)
}

func GunEquipped(o interface{}) bool {
	_, ok := GetEquipment(o.(*Blob).GetGuid(), GunEquipSlot)
	return ok
}

const spawnsPerLevel = 32


func Spawn(name string) *Blob {
	manager := GetManager()
	guid := assemblages[name].MakeEntity(manager)

	return GetBlobs().Get(guid).(*Blob)
}

func SpawnAt(name string, pos geom.Pt2I) (result *Blob) {
	result = Spawn(name)
	PosComp(result.GetGuid()).MoveAbs(pos)
	return
}

func SpawnRandomPos(name string) (result *Blob) {
	return SpawnAt(name, GetSpawnPos())
}

func clearNonplayerEntities() {
	// Bring over player object and player's inventory.
	player := GetBlob(PlayerId())
	keep := make(map[entity.Id]bool)
	keep[player.GetGuid()] = true
	for o := range RecursiveContents(player.GetGuid()).Iter() {
		keep[o.(entity.Id)] = true
	}

	for o := range GetBlobs().EntityComponents().Iter() {
		pair := o.(*entity.IdComponent)
		if _, ok := keep[pair.Entity]; !ok {
			defer GetManager().RemoveEntity(pair.Entity)
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
