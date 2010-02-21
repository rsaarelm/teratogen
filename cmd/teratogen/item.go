package main

import (
	"hyades/entity"
)


const (
	ItemComponent       = entity.ComponentFamily("item")
	MeleeEquipComponent = entity.ComponentFamily("meleeEquip")
	GunEquipComponent   = entity.ComponentFamily("gunEquip")
	ArmorEquipComponent = entity.ComponentFamily("armorEquip")
)

// Equipment slots
const (
	NoEquipSlot = iota
	MeleeEquipSlot
	GunEquipSlot
	ArmorEquipSlot
)


// There's currently no structural difference between item template and item,
// so they can be aliased. This is just a convenience and may change.
type ItemTemplate Item

func (self *ItemTemplate) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return c
}

func (self *ItemTemplate) MakeComponent(manager *entity.Manager, guid entity.Id) {
	result := &Item{
		EquipmentSlot: self.EquipmentSlot,
		Durability: self.Durability,
		WoundBonus: self.WoundBonus,
		DefenseBonus: self.DefenseBonus}
	manager.Handler(ItemComponent).Add(guid, result)
}


type Item struct {
	EquipmentSlot int
	Durability    int
	WoundBonus    int
	DefenseBonus  int
}

func GetItem(id entity.Id) *Item { return GetManager().Handler(ItemComponent).Get(id).(*Item) }

func (self *Item) EquipRelation() *entity.Relation {
	switch self.EquipmentSlot {
	case MeleeEquipSlot:
		return GetManager().Handler(MeleeEquipComponent).(*entity.Relation)
	case GunEquipSlot:
		return GetManager().Handler(GunEquipComponent).(*entity.Relation)
	case ArmorEquipSlot:
		return GetManager().Handler(ArmorEquipComponent).(*entity.Relation)
	}
	return nil
}

func GetEquipment(creature entity.Id, slot string) (guid entity.Id, found bool) {
	// Crashes here if slot isn't a relation component name.
	rel := GetManager().Handler(entity.ComponentFamily(slot)).(*entity.Relation)
	return rel.GetRhs(creature)
}

func SetEquipment(creature entity.Id, slot string, equipment entity.Id) {
	rel := GetManager().Handler(entity.ComponentFamily(slot)).(*entity.Relation)
	rel.AddPair(creature, equipment)
}

// RemoveEquipment remover whatever a creature has equipped in a given slot.
func RemoveEquipment(creature entity.Id, slot string) (removed entity.Id, found bool) {
	rel := GetManager().Handler(entity.ComponentFamily(slot)).(*entity.Relation)
	removed, found = GetEquipment(creature, slot)
	if found {
		rel.RemovePair(creature, removed)
	}
	return
}

// RemoveEquipped removes an item from an equipped relation if it is in one.
func RemoveEquipped(item entity.Id) {
	blob := GetBlobs().Get(item).(*Blob)
	if slot, ok := blob.GetSOpt(PropEquipmentSlot); ok {
		rel := GetManager().Handler(entity.ComponentFamily(slot)).(*entity.Relation)
		rel.RemoveWithRhs(item)
	}
}
