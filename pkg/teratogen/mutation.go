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
)

type mutation struct {
	apply func(id entity.Id)

	// Number of existing mutations before this one can be viable.
	minLevel int

	// Extra predicate for whether the entity can get the mutation. May be nil.
	predicate func(id entity.Id) bool

	// Bit vector of all the mutations that must be present before this one can
	// be viable.
	prereqs uint64
}

var mutations = map[uint64]*mutation{
	MutationStr1:  &mutation{powerMutation, 0, nil, 0},
	MutationStr2:  &mutation{powerMutation, 0, nil, MutationStr1},
	MutationStr3:  &mutation{powerMutation, 0, nil, MutationStr2},
	MutationGrow1: &mutation{growMutation, 0, nil, 0},
	MutationGrow2: &mutation{growMutation, 0, nil, MutationGrow1},
	MutationGrow3: &mutation{growMutation, 0, nil, MutationGrow2},
	MutationEsp:   &mutation{esperMutation, 0, getsEsperMutation, 0},
	MutationTough: &mutation{toughMutation, 0, getsToughMutation, 0},
	// TODO more
}

func GetMutations(id entity.Id) *Mutations {
	if result := GetManager().Handler(MutationsComponent).Get(id); result != nil {
		return result.(*Mutations)
	}
	return nil
}

func Mutate(id entity.Id) {
	crit := GetCreature(id)
	if !canMutate(id) {
		return
	}

	mutations := GetMutations(id)

	// Creatures without a mutations component just go terminal straight away.
	if mutations == nil {
		terminalMutation(id)
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

	mutation := num.RandomChoiceA(iterable.Data(availableMutations(id))).(func(entity.Id))

	EMsg("{sub.Thename} mutate{sub.s}.\n", id, entity.NilId)
	mutation(id)
	crit.Mutations++

	// Heal the creature while at it.
	crit.Wounds = 0

	if crit.Mutations >= terminalMutationCount {
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
	if crit := GetCreature(id); crit != nil {
		return mutationStatusString(crit.Mutations)
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

	if getsGrowMutation(id) {
		vec.Push(growMutation)
	}

	if getsPowerMutation(id) {
		vec.Push(powerMutation)
	}

	if getsEsperMutation(id) {
		vec.Push(esperMutation)
	}

	if getsToughMutation(id) {
		vec.Push(toughMutation)
	}

	return vec
}

func getsGrowMutation(id entity.Id) bool { return true }

func growMutation(id entity.Id) {
	EMsg("{sub.Thename} grow{sub.s} larger.\n", id, entity.NilId)
	GetCreature(id).Scale++
}

func getsPowerMutation(id entity.Id) bool { return true }

func powerMutation(id entity.Id) {
	EMsg("{sub.Thename} grow{sub.s} stronger.\n", id, entity.NilId)
	GetCreature(id).Power++
}

func getsEsperMutation(id entity.Id) bool {
	return !GetCreature(id).HasIntrinsic(IntrinsicEsper)
}

func esperMutation(id entity.Id) {
	EMsg("{sub.Thename} sense{sub.s} minds surrounding {sub.accusative}.\n", id, entity.NilId)
	GetCreature(id).AddIntrinsic(IntrinsicEsper)
}

func getsToughMutation(id entity.Id) bool {
	return !GetCreature(id).HasIntrinsic(IntrinsicTough)
}

func toughMutation(id entity.Id) {
	EMsg("{sub.Thename's} skin hardens into scales.\n", id, entity.NilId)
	GetCreature(id).AddIntrinsic(IntrinsicTough)
}
