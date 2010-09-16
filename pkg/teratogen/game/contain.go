package game

import (
	"exp/iterable"
	"hyades/entity"
)

const ContainComponent = entity.ComponentFamily("contain")

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
	return iterable.Func(func(c chan<- interface{}) {
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

func CountContents(id entity.Id) int { return GetContain().CountRhs(id) }
