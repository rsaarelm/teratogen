package main

import (
	"exp/iterable"
	"fmt"
	"hyades/alg"
	"hyades/dbg"
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

	// Slot type, where does this item go if it's gear.
	PropEquipmentSlot = "equipmentSlot"

	// Equipped gear.
	PropBodyArmorGuid   = "bodyArmorGuid"
	PropMeleeWeaponGuid = "meleeWeaponGuid"
	PropGunWeaponGuid   = "gunWeaponGuid"

	// Extra damage capability
	PropWoundBonus = "woundBonus"
	// Extra defense capability
	PropDefenseBonus = "defenseBonus"

	// How resistant is gear to breaking.
	PropDurability = "durability"

	FlagObstacle = "isObstacle"
)

// Slot type
const (
	SlotBodyArmor = iota
	SlotMeleeWeapon
	SlotGunWeapon
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

type entityPrototype struct {
	Name     string
	Parent   string
	IconId   string
	Class    EntityClass
	Scarcity int
	MinDepth int
	Props    map[string]interface{}
}

func NewPrototype(name, parent, iconId string, class EntityClass, scarcity, minDepth int, a ...) (result *entityPrototype) {
	result = new(entityPrototype)
	result.Name = name
	result.Parent = parent
	result.IconId = iconId
	result.Class = class
	result.Scarcity = scarcity
	result.MinDepth = minDepth

	// Custom settings from varargs.
	props := alg.UnpackEllipsis(a)
	dbg.Assert(len(props)%2 == 0, "NewPrototype: Proplist length is odd.")
	result.Props = make(map[string]interface{})
	for i := 0; i < len(props); i += 2 {
		result.Props[props[i].(string)] = props[i+1]
	}
	return
}

func (self *entityPrototype) applyProps(prototypes map[string]*entityPrototype, target *Entity) {
	if parent, ok := prototypes[self.Parent]; ok {
		parent.applyProps(prototypes, target)
	}
	for key, val := range self.Props {
		target.Set(key, val)
	}
}

func (self *entityPrototype) MakeEntity(prototypes map[string]*entityPrototype, target *Entity) {
	target.IconId = self.IconId
	target.Name = self.Name
	target.Class = self.Class
	self.applyProps(prototypes, target)
}

func (self *entityPrototype) SpawnWeight(depth int) (result float64) {
	const epsilon = 1e-7
	const outOfDepthFactor = 2.0
	scarcity := float64(self.Scarcity)

	if depth < self.MinDepth {
		// Exponentially increase the scarcity for each level out of depth
		outOfDepth := self.MinDepth - depth
		scarcity *= math.Pow(outOfDepthFactor, float64(outOfDepth))
	}

	result = 1.0 / scarcity
	// Make too scarse weights just plain zero.
	if result < epsilon {
		result = 0.0
	}
	return
}

var prototypes = map[string]*entityPrototype{
	// Base prototype for creatures.
	"creature": NewPrototype("creature", "", "", EnemyEntityClass, -1, 0,
		FlagObstacle, 1,
		PropStrength, Fair,
		PropToughness, Fair,
		PropMeleeSkill, Fair,
		PropScale, 0,
		PropWounds, 0,
		PropDensity, 0),
	"protagonist": NewPrototype("protagonist", "creature", "chars:0", PlayerEntityClass, -1, 0,
		PropStrength, Superb,
		PropToughness, Good,
		PropMeleeSkill, Good),
	"zombie": NewPrototype("zombie", "creature", "chars:1", EnemyEntityClass, 100, 0,
		PropStrength, Fair,
		PropToughness, Poor,
		PropMeleeSkill, Fair),
	"dogthing": NewPrototype("dog-thing", "creature", "chars:2", EnemyEntityClass, 150, 0,
		PropStrength, Fair,
		PropToughness, Fair,
		PropMeleeSkill, Good,
		PropScale, -1),
	"ogre": NewPrototype("ogre", "creature", "chars:15", EnemyEntityClass, 600, 5,
		PropStrength, Great,
		PropToughness, Great,
		PropMeleeSkill, Fair,
		PropScale, 3),
	"boss1": NewPrototype("elder spawn", "creature", "chars:5", EnemyEntityClass, 3000, 10,
		PropStrength, Legendary,
		PropToughness, Legendary,
		PropMeleeSkill, Superb,
		PropScale, 5),
	"globe": NewPrototype("health globe", "", "items:1", GlobeEntityClass, 30, 0),
	"plantpot": NewPrototype("plant pot", "", "items:3", ItemEntityClass, 200, 0),
	"pistol": NewPrototype("pistol", "", "items:4", ItemEntityClass, 200, 0,
		PropEquipmentSlot, SlotGunWeapon,
		PropWoundBonus, 1,
		PropDurability, 22),
	"machete": NewPrototype("machete", "", "items:5", ItemEntityClass, 200, 0,
		PropEquipmentSlot, SlotMeleeWeapon,
		PropWoundBonus, 2,
		PropDurability, 100),
	"kevlar": NewPrototype("kevlar armor", "", "items:6", ItemEntityClass, 200, 0,
		PropEquipmentSlot, SlotBodyArmor,
		PropToughness, Good,
		PropDefenseBonus, 1,
		PropDurability, 50),
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
	for o := range world.EntitiesAt(player.GetPos()).Iter() {
		ent := o.(*Entity)
		if ent == player {
			continue
		}
		if ent.GetClass() == GlobeEntityClass {
			// TODO: Different globe effects.
			if player.GetI(PropWounds) > 0 {
				Msg("The globe bursts. You feel better.\n")
				player.Set(PropWounds, player.GetI(PropWounds)-1)
				// Deferring this until the iteration is over.
				defer world.DestroyEntity(ent)
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

	for o := range world.EntitiesAt(target).Iter() {
		ent := o.(*Entity)
		if IsEnemyOf(player, ent) {
			Attack(player, ent)
			return
		}
	}
	// No attack, move normally.
	MovePlayerDir(dir)
	StuffOnGroundMsg()
}

func RunAI() {
	world := GetWorld()
	enemyCount := 0
	for o := range world.Creatures().Iter() {
		crit := o.(*Entity)
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

func TakeableItems(pos geom.Pt2I) iterable.Iterable {
	return iterable.Filter(GetWorld().EntitiesAt(pos), func(o interface{}) bool { return IsTakeableItem(o.(*Entity)) })
}

// TODO: Change other functions to use interface{} instead of *Entity to make
// it easier to use them with iterable functions.

func IsEquippableItem(o interface{}) bool { return o.(*Entity).Has(PropEquipmentSlot) }

func IsCarryingGear(o interface{}) bool {
	return iterable.Any(o.(*Entity).Contents(), IsEquippableItem)
}

func IsCarryingGearFor(o interface{}, slot int) bool {
	return iterable.Any(o.(*Entity).Contents(), func(o interface{}) bool {
		if itemSlot, ok := o.(*Entity).GetIOpt(PropEquipmentSlot); ok && itemSlot == slot {
			return true
		}
		return false
	})
}

func HasContents(o interface{}) bool { return o.(*Entity).GetChild() != nil }

func SmartPlayerPickup() *Entity {
	world := GetWorld()
	player := world.GetPlayer()
	items := iterable.Data(TakeableItems(player.GetPos()))

	if len(items) == 0 {
		Msg("Nothing to take here.\n")
		return nil
	}

	choice := items[0]
	if len(items) > 1 {
		var ok bool
		choice, ok = ObjectChoiceDialog("Pick up which item?", items)
		if !ok {
			Msg("Okay, then.\n")
			return nil
		}
	}
	TakeItem(player, choice.(*Entity))
	return choice.(*Entity)
}

func CanEquipIn(slotId string, e *Entity) bool {
	slot := 0
	switch slotId {
	case PropBodyArmorGuid:
		slot = SlotBodyArmor
	case PropMeleeWeaponGuid:
		slot = SlotMeleeWeapon
	case PropGunWeaponGuid:
		slot = SlotGunWeapon
	default:
		dbg.Die("Unknown equipment slot: %s", slotId)
	}
	if eSlot, ok := e.GetIOpt(PropEquipmentSlot); ok {
		return eSlot == slot
	}
	return false
}
