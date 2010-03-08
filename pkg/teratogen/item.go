package teratogen

import (
	"hyades/entity"
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
