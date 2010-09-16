package game

import (
	"hyades/entity"
)

const FixedInventoryComponent = entity.ComponentFamily("fixedInventory")

const MaxAmmo = 200
const MaxMedKits = 5
const MaxGrenades = 20
const MaxArmor = 100

const AmmoPerClip = 10

// An inventory that holds fixed counts of stuff instead of several different
// items.
type FixedInventory struct {
	Ammo     int
	MedKits  int
	Grenades int
	Armor    int
}

func GetInventory(id entity.Id) *FixedInventory {
	if inv := GetManager().Handler(FixedInventoryComponent).Get(id); inv != nil {
		return inv.(*FixedInventory)
	}
	return nil
}
