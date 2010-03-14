package teratogen

import (
	"container/vector"
	"exp/iterable"
	"hyades/entity"
	"hyades/num"
)

func Mutate(id entity.Id) {
	crit := GetCreature(id)
	if crit.Mutations >= terminalMutationCount {
		// Full chaos, can't mutate further.
		return
	}

	mutation := num.RandomChoiceA(iterable.Data(availableMutations(id))).(func(entity.Id))

	EMsg("{sub.Thename} mutate{sub.s}.\n", id, entity.NilId)
	mutation(id)
	crit.Mutations++

	if crit.Mutations >= terminalMutationCount {
		EMsg("{sub.Thename} mutate{sub.s} further!\n", id, entity.NilId)
		terminalMutation(id)
	}
}

func terminalMutation(id entity.Id) {
	name := GetNameComp(id)
	name.IconId = "chars:20"
	name.Pronoun = PronounIt
	name.Name = "abomination"

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
	GetCreature(id).Traits |= IntrinsicEsper
}

func getsToughMutation(id entity.Id) bool {
	return !GetCreature(id).HasIntrinsic(IntrinsicTough)
}

func toughMutation(id entity.Id) {
	EMsg("{sub.Thename's} skin hardens into scales.\n", id, entity.NilId)
	GetCreature(id).Traits |= IntrinsicTough
}
