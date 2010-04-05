package teratogen

import (
	"container/vector"
	"exp/iterable"
	"hyades/entity"
	"hyades/num"
)

const MutationsComponent = entity.ComponentFamily("mutations")

type Mutations struct {
	mutations uint64
}

const (
	MutationStr1 = 1 << iota
	MutationStr2
	MutationStr3
	MutationGrow1
	MutationGrow2
	MutationGrow3
	MutationEsp
	MutationTough
	MutationShimmer
	MutationHooves
	MutationHorns
)

type mutation struct {
	flag uint64

	apply func(id entity.Id)

	// Number of existing mutations before this one can be viable.
	minLevel int

	// Extra predicate for whether the entity can get the mutation. May be nil.
	predicate func(id entity.Id) bool

	// Bit vector of all the mutations that must be present before this one can
	// be viable.
	prereqs uint64
}

func (self *mutation) Apply(id entity.Id) {
	mutations := GetMutations(id)
	if mutations == nil {
		return
	}

	mutations.mutations |= self.flag

	self.apply(id)
}

func (self *mutation) CanApplyTo(id entity.Id) bool {
	mutations := GetMutations(id)
	if mutations == nil {
		return false
	}
	level := mutations.MutationLevel()

	if mutations.mutations&self.flag != 0 {
		// This mutation is already present.
		return false
	}

	if self.minLevel > level {
		// Level too low.
		return false
	}
	if mutations.mutations&self.prereqs != self.prereqs {
		// Unmatched prerequisites.
		return false
	}
	if self.predicate != nil && !self.predicate(id) {
		// Special predicate didn't match.
		return false
	}

	return true
}

var mutations = map[uint64]*mutation{
	MutationStr1:    &mutation{MutationStr1, powerMutation, 0, nil, 0},
	MutationStr2:    &mutation{MutationStr2, powerMutation, 0, nil, MutationStr1},
	MutationStr3:    &mutation{MutationStr3, powerMutation, 0, nil, MutationStr2},
	MutationGrow1:   &mutation{MutationGrow1, growMutation, 0, nil, 0},
	MutationGrow2:   &mutation{MutationGrow2, growMutation, 0, nil, MutationGrow1},
	MutationGrow3:   &mutation{MutationGrow3, growMutation, 0, nil, MutationGrow2},
	MutationEsp:     &mutation{MutationEsp, esperMutation, 0, hasIntrinsicFilter(IntrinsicEsper), 0},
	MutationTough:   &mutation{MutationTough, toughMutation, 0, hasIntrinsicFilter(IntrinsicTough), 0},
	MutationShimmer: &mutation{MutationShimmer, shimmerMutation, 0, hasIntrinsicFilter(IntrinsicShimmer), 0},
	MutationHorns:   &mutation{MutationHorns, hornsMutation, 0, hasIntrinsicFilter(IntrinsicHorns), 0},
	MutationHooves:  &mutation{MutationHooves, hoovesMutation, 0, hasIntrinsicFilter(IntrinsicHooves), 0},
	// TODO more
}

func GetMutations(id entity.Id) *Mutations {
	if result := GetManager().Handler(MutationsComponent).Get(id); result != nil {
		return result.(*Mutations)
	}
	return nil
}

func (self *Mutations) MutationLevel() int {
	return num.NumberOfSetBitsU64(self.mutations)
}

func (self *Mutations) HasMutation(mutation uint64) bool {
	return self.mutations&mutation != 0
}

func (self *Mutations) SetMutation(mutation uint64) {
	self.mutations |= mutation
}

func Mutate(id entity.Id) {
	crit := GetCreature(id)
	if !canMutate(id) {
		return
	}

	if crit.HasStatus(StatusMutationShield) {
		EMsg("{sub.Thename} remain{sub.s} unchanged.\n", id, entity.NilId)
		if num.OneChanceIn(2) {
			crit.RemoveStatus(StatusMutationShield)
			if id == PlayerId() {
				Msg("You feel less stable.\n")
			}
		}
		return
	}

	mutations := GetMutations(id)

	// Creatures without a mutations component just go terminal straight away.
	if mutations == nil {
		terminalMutation(id)
		return
	}

	available := iterable.Data(availableMutations(id))

	if len(available) == 0 {
		// No available mutations.
		EMsg("{sub.Thename} feel{sub.s} unstable for a moment.\n", id, entity.NilId)
		return
	}

	mut := num.RandomChoiceA(available).(*mutation)

	EMsg("{sub.Thename} mutate{sub.s}.\n", id, entity.NilId)
	mut.Apply(id)

	// Heal the creature while at it.
	crit.Wounds = 0

	if mutations.MutationLevel() >= terminalMutationCount {
		EMsg("{sub.Thename} mutate{sub.s} further!\n", id, entity.NilId)
		terminalMutation(id)
	}
}

func canMutate(id entity.Id) bool {
	// Chaos creatures won't mutate, others do.
	return !GetCreature(id).HasIntrinsic(IntrinsicChaosSpawn)
}

func terminalMutation(id entity.Id) {
	name := GetNameComp(id)
	name.IconId = "chars:20"
	name.Pronoun = PronounIt
	name.Name = "abomination"

	crit := GetCreature(id)
	crit.AddIntrinsic(IntrinsicChaosSpawn)

	// Scramble base stats.
	crit.Scale += FudgeDice()
	crit.Power += FudgeDice()
	crit.Skill += FudgeDice()

	if id == PlayerId() {
		GameOver("became one with the Tau wave.")
	}
}

const terminalMutationCount = 10

func MutationStatus(id entity.Id) string {
	if mut := GetMutations(id); mut != nil {
		return mutationStatusString(mut.MutationLevel())
	}
	return ""
}

func mutationStatusString(numMutations int) string {
	switch numMutations {
	case 1:
		return "touched"
	case 2:
		return "tainted"
	case 3:
		return "unclean"
	case 4:
		return "altered"
	case 5:
		return "warped"
	case 6:
		return "blighted"
	case 7:
		return "noxious"
	case 8:
		return "corrupt"
	case 9:
		return "baneful"
	case 10:
		return "forsaken"
	}
	return ""
}

func availableMutations(id entity.Id) iterable.Iterable {
	vec := new(vector.Vector)

	for _, m := range mutations {
		if m.CanApplyTo(id) {
			vec.Push(m)
		}
	}

	return vec
}

func growMutation(id entity.Id) {
	EMsg("{sub.Thename} grow{sub.s} larger.\n", id, entity.NilId)
	GetCreature(id).Scale++
}

func powerMutation(id entity.Id) {
	EMsg("{sub.Thename} grow{sub.s} stronger.\n", id, entity.NilId)
	GetCreature(id).Power++
}

func hasIntrinsicFilter(intrinsic int32) func(id entity.Id) bool {
	return func(id entity.Id) bool {
		return !GetCreature(id).HasIntrinsic(intrinsic)
	}
}

func esperMutation(id entity.Id) {
	EMsg("{sub.Thename} sense{sub.s} minds surrounding {sub.accusative}.\n", id, entity.NilId)
	GetCreature(id).AddIntrinsic(IntrinsicEsper)
}

func toughMutation(id entity.Id) {
	EMsg("{sub.Thename's} skin hardens into scales.\n", id, entity.NilId)
	GetCreature(id).AddIntrinsic(IntrinsicTough)
}

func shimmerMutation(id entity.Id) {
	EMsg("{sub.Thename} begins to shimmer in and out of reality.\n", id, entity.NilId)
	GetCreature(id).AddIntrinsic(IntrinsicShimmer)
}

func hornsMutation(id entity.Id) {
	EMsg("Horns grow into {sub.thename's} head.\n", id, entity.NilId)
	GetCreature(id).AddIntrinsic(IntrinsicHorns)
}

func hoovesMutation(id entity.Id) {
	EMsg("{sub.Thename's} feet deform into cloven hooves.\n", id, entity.NilId)
	GetCreature(id).AddIntrinsic(IntrinsicHooves)
}
