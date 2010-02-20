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

	// Slot type, where does this item go if it's gear.
	PropEquipmentSlot = "equipmentSlot"

	// Equipped gear.
	PropBodyArmorGuid   = "armorEquip"
	PropMeleeWeaponGuid = "meleeEquip"
	PropGunWeaponGuid   = "gunEquip"

	// Extra damage capability
	PropWoundBonus = "woundBonus"
	// Extra defense capability
	PropDefenseBonus = "defenseBonus"

	// How resistant is gear to breaking.
	PropDurability = "durability"

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
	filter := func(o interface{}) bool { return IsEnemyOf(id, o.(*Blob).GetGuid()) }

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

func Attack(attacker *Blob, defender *Blob) {
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
		DamageEquipment(attacker, PropMeleeWeaponGuid)
		DamageEquipment(defender, PropBodyArmorGuid)
	} else {
		Msg("%v missed.\n", txt.Capitalize(attacker.GetName()))
	}
}

func Shoot(attacker *Blob, target geom.Pt2I) {
	if !GunEquipped(attacker) {
		return
	}

	// TODO: Aiming precision etc.
	origin := attacker.GetPos()
	var hitPos geom.Pt2I
	for o := range iterable.Drop(geom.Line(origin, target), 1).Iter() {
		hitPos = o.(geom.Pt2I)
		if !IsOpen(hitPos) {
			break
		}
	}

	damageFactor := 0
	if gun, ok := GetEquipment(attacker.GetGuid(), PropGunWeaponGuid); ok {
		damageFactor += GetBlobs().Get(gun).(*Blob).GetI(PropWoundBonus)
	}

	p1, p2 := draw.Pt(Tile2WorldPos(GetPos(PlayerId()))), draw.Pt(Tile2WorldPos(hitPos))
	go LineAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), p1, p2, 2e8, gfx.White, gfx.DarkRed, config.Scale*config.TileScale)

	// TODO: Sparks when hitting walls.
	DamagePos(hitPos, damageFactor, attacker)

	DamageEquipment(attacker, PropGunWeaponGuid)
}

func DamageEquipment(ent *Blob, slot string) {
	if guid, ok := GetEquipment(ent.GetGuid(), slot); ok {
		o := GetBlobs().Get(guid).(*Blob)
		if num.OneChanceIn(o.GetI(PropDurability)) {
			if slot == PropGunWeaponGuid {
				Msg("The %s's %s is out of ammo.\n", ent.GetName(), o.GetName())
			} else {
				Msg("The %s's %s breaks.\n", ent.GetName(), o.GetName())
			}
			DestroyBlob(o)
		}
	}
}

func DamagePos(pos geom.Pt2I, woundLevel int, cause *Blob) {
	for o := range iterable.Filter(EntitiesAt(pos), IsCreature).Iter() {
		o.(*Blob).Damage(woundLevel, cause)
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
	player.TryMove(geom.Dir8ToVec(dir))

	// TODO: More general collision code, do collisions for AI creatures
	// too.

	// See if the player collided with something fun.
	for o := range EntitiesAt(player.GetPos()).Iter() {
		ent := o.(*Blob)
		if ent == player {
			continue
		}
		if ent.GetClass() == GlobeEntityClass {
			// TODO: Different globe effects.
			if player.GetI(PropWounds) > 0 {
				Msg("The globe bursts. You feel better.\n")
				PlaySound("heal")
				player.Set(PropWounds, player.GetI(PropWounds)-1)
				// Deferring this until the iteration is over.
				defer DestroyBlob(ent)
			}
		}
	}

	GetLos().DoLos(player.GetPos())
}

func SmartMovePlayer(dir int) {
	vec := geom.Dir8ToVec(dir)
	target := GetPos(PlayerId()).Plus(vec)

	for o := range EnemiesAt(PlayerId(), target).Iter() {
		Attack(GetBlob(PlayerId()), o.(*Blob))
		return
	}
	// No attack, move normally.
	MovePlayerDir(dir)
	StuffOnGroundMsg()
}

func RunAI() {
	enemyCount := 0
	for o := range Creatures().Iter() {
		crit := o.(*Blob)
		if crit.GetGuid() != PlayerId() {
			enemyCount++
		}
		DoAI(crit)
	}
}

func GameOver(reason string) {
	MsgMore()
	fmt.Printf("%v %v\n", txt.Capitalize(GetBlob(PlayerId()).Name), reason)
	Quit()
}

// Return whether the entity moves around by itself and shouldn't be shown in
// map memory.
func IsMobile(entity *Blob) bool { return entity.GetClass() > CreatureEntityClassStartMarker }

func PlayerEnterStairs() {
	if GetArea().GetTerrain(GetPos(PlayerId())) == TerrainStairDown {
		Msg("Going down...\n")
		NextLevel()
	} else {
		Msg("There are no stairs here.\n")
	}
}

func NextLevel() { GetContext().EnterLevel(GetCurrentLevel()) }

func IsCreature(o interface{}) bool {
	switch o.(*Blob).GetClass() {
	case PlayerEntityClass, EnemyEntityClass:
		return true
	}
	return false
}

func IsTakeableItem(e *Blob) bool { return e.Class == ItemEntityClass }

func TakeItem(subject *Blob, item *Blob) {
	item.InsertSelf(subject)
	Msg("%v takes %v.\n", txt.Capitalize(subject.GetName()), item.GetName())
}

func DropItem(subject *Blob, item *Blob) {
	// TODO: Check if the subject is holding the item.
	item.RemoveSelf()
	item.MoveAbs(subject.GetPos())
	Msg("%v drops %v.\n", txt.Capitalize(subject.GetName()), item.GetName())
}

func TakeableItems(pos geom.Pt2I) iterable.Iterable {
	return iterable.Filter(EntitiesAt(pos), func(o interface{}) bool { return IsTakeableItem(o.(*Blob)) })
}

// TODO: Change other functions to use interface{} instead of *Blob to make
// it easier to use them with iterable functions.

func IsEquippableItem(o interface{}) bool { return o.(*Blob).Has(PropEquipmentSlot) }

func IsCarryingGear(o interface{}) bool {
	return iterable.Any(o.(*Blob).Contents(), IsEquippableItem)
}

func IsCarryingGearFor(o interface{}, slot string) bool {
	return iterable.Any(o.(*Blob).Contents(), func(o interface{}) bool {
		if itemSlot, ok := o.(*Blob).GetSOpt(PropEquipmentSlot); ok && itemSlot == slot {
			return true
		}
		return false
	})
}

func HasContents(o interface{}) bool { return o.(*Blob).HasContents() }

func IsUsable(o interface{}) bool { return o.(*Blob).Has(PropItemUse) }

func HasUsableItems(o interface{}) bool { return iterable.Any(o.(*Blob).Contents(), IsUsable) }

func UseItem(user *Blob, item *Blob) {
	if use, ok := item.GetIOpt(PropItemUse); ok {
		switch use {
		case NoUse:
			Msg("Nothing happens.\n")
		case MedkitUse:
			if user.GetI(PropWounds) > 0 {
				Msg("You feel much better.\n")
				PlaySound("heal")
				user.Set(PropWounds, 0)
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
	items := iterable.Data(TakeableItems(player.GetPos()))

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
	slot, ok := item.GetSOpt(PropEquipmentSlot)
	if !ok {
		return
	}
	if _, ok := owner.GetGuidOpt(slot); ok {
		// Already got something equipped.
		return
	}
	SetEquipment(owner.GetGuid(), slot, item.GetGuid())
	Msg("Equipped %v.\n", item)
}

func CanEquipIn(slotId string, e *Blob) bool {
	if eSlot, ok := e.GetSOpt(PropEquipmentSlot); ok {
		return eSlot == slotId
	}
	return false
}

func EntityDist(o1, o2 interface{}) float64 {
	e1, ok1 := o1.(interface {
		GetPos() geom.Pt2I
	})
	e2, ok2 := o2.(interface {
		GetPos() geom.Pt2I
	})
	if !ok1 || !ok2 {
		return math.MaxFloat64
	}
	return e1.GetPos().Minus(e2.GetPos()).Abs()
}

func CreaturesSeenBy(o interface{}) iterable.Iterable {
	ent := o.(*Blob)
	pred := func(o interface{}) bool { return ent.CanSeeTo(o.(*Blob).GetPos()) }
	return iterable.Filter(OtherCreatures(o), pred)
}

func ClosestCreatureSeenBy(o interface{}) *Blob {
	distFromSelf := func(o1 interface{}) float64 { return EntityDist(o1, o) }
	ret, ok := alg.IterMin(CreaturesSeenBy(o), distFromSelf)
	if !ok {
		return nil
	}
	return ret.(*Blob)
}

func GunEquipped(o interface{}) bool {
	_, ok := GetEquipment(o.(*Blob).GetGuid(), PropGunWeaponGuid)
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
	result.MoveAbs(pos)
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
	for ent := range player.RecursiveContents().Iter() {
		keep[ent.(*Blob).GetGuid()] = true
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
		proto := assemblages[item.(string)][BlobComponent].(*blobTemplate)
		return SpawnWeight(proto.Scarcity, proto.MinDepth, depth)
	}
	values := make([]interface{}, len(assemblages))
	i := 0
	for name, _ := range assemblages {
		values[i] = name
		i++
	}
	return num.MakeWeightedDist(weightFn, values)
}
