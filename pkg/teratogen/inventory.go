package teratogen

import (
	"hyades/entity"
)

const FixedInventoryComponent = entity.ComponentFamily("fixedInventory")

// An inventory that holds fixed counts of stuff instead of several different
// items.
type FixedInventory struct {
	Ammo     int
	MedKits  int
	Grenades int
	Armor    int
}

func GetInventory(id entity.Id) *FixedInventory {
	if inv := GetManager().Handler(CreatureComponent).Get(id); inv != nil {
		return inv.(*FixedInventory)
	}
	return nil
}
