package main

import (
	"hyades/entity"
)


const ItemComponent = entity.ComponentFamily("item")


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
