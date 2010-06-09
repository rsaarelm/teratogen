package teratogen

import (
	"exp/iterable"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
)


const (
	ItemComponent       = entity.ComponentFamily("item")
	MeleeEquipComponent = entity.ComponentFamily("meleeEquip")
	GunEquipComponent   = entity.ComponentFamily("gunEquip")
	ArmorEquipComponent = entity.ComponentFamily("armorEquip")
)

type EquipSlot int

// Equipment slots
const (
	NoEquipSlot EquipSlot = iota
	MeleeEquipSlot
	GunEquipSlot
	ArmorEquipSlot
)


// Item use type
type ItemUse int

const (
	NoUse ItemUse = iota
	MedkitUse
	StabilizerUse
)

const (
	NoItemTrait = 1 << iota
	ItemRapidFire
	ItemKnockback
	// TODO: Hard suit effect: Power up character strength and resilience when worn.
	ItemHardsuit
)

// How big is the knockback effect.
const ItemKnockbackStrength = 3


func (self EquipSlot) Relation() *entity.Relation {
	switch self {
	case MeleeEquipSlot:
		return GetManager().Handler(MeleeEquipComponent).(*entity.Relation)
	case GunEquipSlot:
		return GetManager().Handler(GunEquipComponent).(*entity.Relation)
	case ArmorEquipSlot:
		return GetManager().Handler(ArmorEquipComponent).(*entity.Relation)
	}
	return nil
}


// There's currently no structural difference between item template and item,
// so they can be aliased. This is just a convenience and may change.
type ItemTemplate Item

func (self *ItemTemplate) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return c
}

func (self *ItemTemplate) MakeComponent(manager *entity.Manager, guid entity.Id) {
	result := &Item{
		EquipmentSlot: self.EquipmentSlot,
		Durability:    self.Durability,
		WoundBonus:    self.WoundBonus,
		DefenseBonus:  self.DefenseBonus,
		Use:           self.Use,
		Traits:        self.Traits}
	manager.Handler(ItemComponent).Add(guid, result)
}


type Item struct {
	EquipmentSlot EquipSlot
	Durability    int
	WoundBonus    int
	DefenseBonus  int
	Use           ItemUse
	Traits        int32
}

func GetItem(id entity.Id) *Item {
	if id := GetManager().Handler(ItemComponent).Get(id); id != nil {
		return id.(*Item)
	}
	return nil
}

func IsItem(id entity.Id) bool { return GetItem(id) != nil }

func (self *Item) HasTrait(trait int32) bool { return self.Traits&trait != 0 }

func GetEquipment(creature entity.Id, slot EquipSlot) (guid entity.Id, found bool) {
	if rel := slot.Relation(); rel != nil {
		return rel.GetRhs(creature)
	}
	return
}

func SetEquipment(creature entity.Id, slot EquipSlot, equipment entity.Id) {
	if rel := slot.Relation(); rel != nil {
		rel.AddPair(creature, equipment)
	}
}

// RemoveEquipment remover whatever a creature has equipped in a given slot.
func RemoveEquipment(creature entity.Id, slot EquipSlot) (removed entity.Id, found bool) {
	if rel := slot.Relation(); rel != nil {
		removed, found = GetEquipment(creature, slot)
		if found {
			rel.RemovePair(creature, removed)
		}
	}
	return
}

// RemoveEquipped removes an item from an equipped relation if it is in one.
func RemoveEquipped(itemId entity.Id) {
	if item := GetItem(itemId); item != nil {
		if item.EquipmentSlot != NoEquipSlot {
			item.EquipmentSlot.Relation().RemoveWithRhs(itemId)
		}
	}
}

func CanEquipIn(slot EquipSlot, itemId entity.Id) bool {
	item := GetItem(itemId)
	return slot != NoEquipSlot && item != nil && slot == item.EquipmentSlot
}

func IsTakeableItem(e entity.Id) bool { return IsItem(e) }

func CarryLimit(id entity.Id) int { return 8 }

func TakeItem(takerId, itemId entity.Id) bool {
	numCarrying := CountContents(takerId)
	if numCarrying < CarryLimit(takerId) {
		SetParent(itemId, takerId)
		EMsg("{sub.Thename} take{sub.s} {obj.thename}.\n", takerId, itemId)
		return true
	} else {
		EMsg("{sub.Thename} can't carry any more items.\n", takerId, itemId)
	}

	return false
}

func DropItem(dropperId, itemId entity.Id) {
	SetParent(itemId, entity.NilId)
	EMsg("{sub.Thename} drop{sub.s} {obj.thename}.\n", dropperId, itemId)
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
			if crit.Health < 1.0 {
				EMsg("{sub.Thename} feel{sub.s} much better.\n", userId, itemId)
				Fx().Heal(userId, 1)
				crit.Health = 1.0
				Destroy(itemId)
			} else {
				EMsg("{sub.Thename} feel{sub.s} fine already.\n", userId, itemId)
			}
		case StabilizerUse:
			crit := GetCreature(userId)
			if !crit.HasStatus(StatusMutationShield) {
				EMsg("{sub.Thename} feel{sub.s} stable.\n", userId, itemId)
			} else {
				EMsg("{sub.Thename} inject{sub.s} the liquid.\n", userId, itemId)
			}
			crit.AddStatus(StatusMutationShield)
			Destroy(itemId)
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
	EMsg("{sub.Thename} equip{sub.s} {obj.thename}.\n", ownerId, itemId)
}

func GunEquipped(id entity.Id) bool {
	_, ok := GetEquipment(id, GunEquipSlot)
	return ok
}

func RapidFireGunEquipped(id entity.Id) bool {
	gunId, ok := GetEquipment(id, GunEquipSlot)
	if !ok {
		return false
	}

	if gun := GetItem(gunId); gun != nil {
		return gun.HasTrait(ItemRapidFire)
	}

	return false
}
