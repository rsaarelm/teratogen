package entity

import (
	"exp/iterable"
	"hyades/alg"
	"hyades/mem"
	"io"
)

type ModelConstraint int

const (
	OneToOne   ModelConstraint = 0
	OneToMany  ModelConstraint = 1
	ManyToOne  ModelConstraint = 2
	ManyToMany ModelConstraint = 3
)

type Pair struct {
	Lhs, Rhs Id
}

// Relation is a two-place relation between entities. Entities on each side of
// the relation can be constrained to occur only once per relation, turning
// the relation from many-to-many to one-to-many, many-to-one or one-to-one.
type Relation struct {
	lhsIndex        map[Id](map[Id]bool)
	rhsIndex        map[Id](map[Id]bool)
	modelConstraint ModelConstraint
}

func NewRelation(constraint ModelConstraint) *Relation {
	return new(Relation).Init(constraint)
}

func (self *Relation) lhsConstrained() bool { return self.modelConstraint&1 == 0 }

func (self *Relation) rhsConstrained() bool { return self.modelConstraint&2 == 0 }

func (self *Relation) Init(constraint ModelConstraint) *Relation {
	self.lhsIndex = make(map[Id](map[Id]bool))
	self.rhsIndex = make(map[Id](map[Id]bool))
	self.modelConstraint = constraint
	return self
}

// Clear empties the relation of all data.
func (self *Relation) Clear() { self.Init(self.modelConstraint) }

// AddPair adds pair <lhs, rhs> to the relation. If the relation's model
// constraint states that an entity may occur only once on the lhs, any
// previous lhs occurrences of the entity will be removed. Constraint to rhs
// values is handled similarly.
func (self *Relation) AddPair(lhs, rhs Id) {
	if self.lhsConstrained() {
		self.lhsIndex[lhs] = nil, false
	}
	if self.rhsConstrained() {
		self.rhsIndex[rhs] = nil, false
	}

	if _, ok := self.lhsIndex[lhs]; !ok {
		self.lhsIndex[lhs] = make(map[Id]bool)
	}

	if _, ok := self.rhsIndex[lhs]; !ok {
		self.rhsIndex[lhs] = make(map[Id]bool)
	}

	self.lhsIndex[lhs][rhs] = true
	self.rhsIndex[rhs][lhs] = true
}

// RemovePair removes the pair <lhs, rhs> from the relation, if it is present.
func (self *Relation) RemovePair(lhs, rhs Id) {
	if lhsMap, ok := self.lhsIndex[lhs]; ok {
		lhsMap[rhs] = false, false
		// Remove empty map
		if len(lhsMap) == 0 {
			self.lhsIndex[lhs] = nil, false
		}
	}

	if rhsMap, ok := self.rhsIndex[rhs]; ok {
		rhsMap[lhs] = false, false
		// Remove empty map
		if len(rhsMap) == 0 {
			self.rhsIndex[rhs] = nil, false
		}
	}
}

// RemoveWithLhs removes all pairs <lhs, *> with the given left-hand side
// value.
func (self *Relation) RemoveWithLhs(lhs Id) {
	if rhsMap, ok := self.lhsIndex[lhs]; ok {
		for rhs, _ := range rhsMap {
			// Defer the removal to keep map modification from messing the
			// iteration.
			defer self.RemovePair(lhs, rhs)
		}
	}
}

// RemoveWithRhs removes all pairs <*, rhs> with the given right-hand side
// value.
func (self *Relation) RemoveWithRhs(rhs Id) {
	if lhsMap, ok := self.rhsIndex[rhs]; ok {
		for lhs, _ := range lhsMap {
			// Defer the removal to keep map modification from messing the
			// iteration.
			defer self.RemovePair(lhs, rhs)
		}
	}
}

// RemoveAllWith removes all pairs with the given entity on either rhs or lhs.
func (self *Relation) RemoveAllWith(entity Id) {
	for o := range self.IterPairs().Iter() {
		pair := o.(Pair)
		if pair.Lhs == entity || pair.Rhs == entity {
			defer self.RemovePair(pair.Lhs, pair.Rhs)
		}
	}
}

// HasPair returns whether the pair <lhs, rhs> is in the relation.
func (self *Relation) HasPair(lhs, rhs Id) bool {
	if lhsMap, ok := self.lhsIndex[lhs]; ok {
		_, ok2 := lhsMap[rhs]
		return ok2
	}
	return false
}

// IterRhs iterates all right-hand side values for the given left-hand side.
func (self *Relation) IterRhs(lhs Id) iterable.Iterable {
	return alg.IterFunc(func(c chan<- interface{}) {
		if rhsMap, ok := self.lhsIndex[lhs]; ok {
			for rhs, _ := range rhsMap {
				c <- rhs
			}
		}
		close(c)
	})
}

// IterLhs iterates all left-hand side values for the given right-hand side.
func (self *Relation) IterLhs(rhs Id) iterable.Iterable {
	return alg.IterFunc(func(c chan<- interface{}) {
		if lhsMap, ok := self.rhsIndex[rhs]; ok {
			for lhs, _ := range lhsMap {
				c <- lhs
			}
		}
		close(c)
	})
}

// GetRhs gets an arbitrary right-hand side value for the given left-hand
// value, if there are any. Useful for ManyToOne relations, where there can't
// be more than one rhs value for any lhs value.
func (self *Relation) GetRhs(lhs Id) (rhs Id, found bool) {
	found = false
	if m, ok := self.lhsIndex[lhs]; ok {
		for val, _ := range m {
			rhs = val
			found = true
			return
		}
	}
	return
}

// GetLhs gets an arbitrary left-hand side value for the given right-hand
// value, if there are any. Useful for OneToMany relations, where there can't
// be more than one lhs value for any rhs value.
func (self *Relation) GetLhs(rhs Id) (lhs Id, found bool) {
	found = false
	if m, ok := self.rhsIndex[rhs]; ok {
		for val, _ := range m {
			lhs = val
			found = true
			return
		}
	}
	return
}

// IterPairs iterates through all the pairs in the relation. It yields Pair
// structs.
func (self *Relation) IterPairs() iterable.Iterable {
	return alg.IterFunc(func(c chan<- interface{}) {
		for lhs, lhsMap := range self.lhsIndex {
			for rhs, _ := range lhsMap {
				c <- Pair{lhs, rhs}
			}
		}
		close(c)
	})
}

func (self *Relation) Add(guid Id, component interface{}) {
	// No-op.
}

func (self *Relation) Get(guid Id) interface{} {
	// No-op.
	return nil
}

func (self *Relation) Remove(guid Id) { self.RemoveAllWith(guid) }

func (self *Relation) EntityComponents() iterable.Iterable {
	return alg.EmptyIter()
}

func (self *Relation) Serialize(out io.Writer) {
	mem.WriteFixed(out, byte(self.modelConstraint))
	pairs := iterable.Data(self.IterPairs())
	mem.WriteNTimes(out, len(pairs), func(i int, out io.Writer) {
		pair := pairs[i].(Pair)
		mem.WriteFixed(out, int64(pair.Lhs))
		mem.WriteFixed(out, int64(pair.Rhs))
	})
}

func (self *Relation) Deserialize(in io.Reader) {
	constraint := ModelConstraint(mem.ReadByte(in))
	self.Init(constraint)
	mem.ReadNTimes(in,
		func(count int) {},
		func(i int, in io.Reader) {
			lhs, rhs := Id(mem.ReadInt64(in)), Id(mem.ReadInt64(in))
			self.AddPair(lhs, rhs)
		})
}
