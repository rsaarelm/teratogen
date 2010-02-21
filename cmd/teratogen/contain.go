package main

import (
	"exp/iterable"
	"hyades/alg"
	"hyades/entity"
)

const (
	ContainComponent    = entity.ComponentFamily("contain")
	MeleeEquipComponent = entity.ComponentFamily("meleeEquip")
	GunEquipComponent   = entity.ComponentFamily("gunEquip")
	ArmorEquipComponent = entity.ComponentFamily("armorEquip")
)

// GetContain gets the containment relation. Lhs is the container and rhs
// values are the immediate containees (transitive containment, being in a
// container in a container, won't show up in the top relation).
func GetContain() *entity.Relation {
	return GetManager().Handler(ContainComponent).(*entity.Relation)
}

func GetParent(id entity.Id) entity.Id {
	parentId, ok := GetContain().GetLhs(id)
	if ok {
		return parentId
	}
	return entity.NilId
}

func GetTopParent(id entity.Id) (topId entity.Id) {
	for parentId := GetParent(id); parentId != entity.NilId; parentId = GetParent(parentId) {
		topId = parentId
	}
	return
}

func SetParent(id, newParentId entity.Id) {
	oldParentId := GetParent(id)

	if newParentId == oldParentId {
		return
	}

	if newParentId == entity.NilId {
		if parentPos, ok := GetParentPosOrPos(id); oldParentId != entity.NilId && ok {
			// Move to the position of the topmost positioned parent when
			// removing from containment.
			SetPos(id, parentPos)
		}
	}

	GetContain().RemoveWithRhs(id)
	RemoveEquipped(id)

	if newParentId != entity.NilId {
		GetContain().AddPair(newParentId, id)
	}
}

// Contents iterates through the children but not the grandchildren of the
// entity.
func Contents(id entity.Id) iterable.Iterable { return GetContain().IterRhs(id) }

// RecursiveContents iterates through all children and grandchildren of the
// entity.
func RecursiveContents(id entity.Id) iterable.Iterable {
	return alg.IterFunc(func(c chan<- interface{}) {
		for o := range Contents(id).Iter() {
			c <- o
			for q := range RecursiveContents(o.(entity.Id)).Iter() {
				c <- q
			}
		}
		close(c)
	})
}

func HasContents(id entity.Id) bool {
	_, ok := GetContain().GetRhs(id)
	return ok
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
