package game

import (
	"container/vector"
	"hyades/entity"
	"hyades/num"
	"math"
	"rand"
)

const MutationsComponent = entity.ComponentFamily("mutations")

const ExpScale float64 = 0.0002
const LevelUpHpBonus = 10

const numLevels = 20

type Mutations struct {
	mutations uint64
	Humanity  float64
}

const (
	MutationEsp = 1 << iota
	MutationTough
	MutationShimmer
	MutationHooves
	MutationHorns
	MutationFast
	MutationCryo
	MutationLightning
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
	MutationEsp:       &mutation{MutationEsp, esperMutation, 0, hasIntrinsicFilter(IntrinsicEsper), 0},
	MutationTough:     &mutation{MutationTough, toughMutation, 2, nil, 0},
	MutationShimmer:   &mutation{MutationShimmer, shimmerMutation, 4, hasIntrinsicFilter(IntrinsicShimmer), 0},
	MutationHorns:     &mutation{MutationHorns, hornsMutation, 0, hasIntrinsicFilter(IntrinsicHorns), 0},
	MutationHooves:    &mutation{MutationHooves, hoovesMutation, 0, hasIntrinsicFilter(IntrinsicHooves), 0},
	MutationFast:      &mutation{MutationFast, fastMutation, 6, hasIntrinsicFilter(IntrinsicFast), 0},
	MutationCryo:      &mutation{MutationCryo, cryoMutation, 0, nil, 0},
	MutationLightning: &mutation{MutationLightning, lightningMutation, 0, nil, 0},
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

func (self *Mutations) HumanityLevel() float64 {
	return self.Humanity
}

func (self *Mutations) GiveExp(id entity.Id, amount float64) {
	oldHumanity := self.Humanity
	self.Humanity -= amount * ExpScale

	// XXX: Subtract the very small number from both to keep a mutation from
	// happening at the very first humanity drop from 100 % to 99.whatever %.
	for i := int(math.Floor(self.Humanity*numLevels - 1e-9)); i < int(math.Floor(oldHumanity*numLevels-1e-9)); i++ {
		Mutate(id)
	}
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

	mutations := GetMutations(id)

	// Creatures without a mutations component just go terminal straight away.
	if mutations == nil {
		terminalMutation(id)
		return
	}

	crit.healthScale += LevelUpHpBonus

	// Level-up full heal
	crit.Health = 1.0

	available := ([]interface{})(*availableMutations(id))

	if len(available) == 0 {
		// No available mutations.
		EMsg("{sub.Thename} feel{sub.s} unstable for a moment.\n", id, entity.NilId)
	} else {
		mut := num.RandomChoiceA(available).(*mutation)

		EMsg("{sub.Thename} mutate{sub.s}.\n", id, entity.NilId)
		mut.Apply(id)

		if id == PlayerId() {
			Fx().MorePrompt()
		}
	}

	if mutations.HumanityLevel() <= 0 {
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

	if id == PlayerId() {
		GameOver("became one with the Tau wave.")
	}
}

func availableMutations(id entity.Id) *vector.Vector {
	vec := new(vector.Vector)

	for _, m := range mutations {
		if m.CanApplyTo(id) {
			vec.Push(m)
		}
	}

	return vec
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
	GetCreature(id).healthScale *= 1.2
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

func fastMutation(id entity.Id) {
	EMsg("{sub.Thename} {sub.is} suddenly moving faster.\n", id, entity.NilId)
	GetCreature(id).AddIntrinsic(IntrinsicFast)
}

func addPower(id entity.Id, power PowerId) {
	crit := GetCreature(id)
	// Put power in an empty slot if there is one, otherwise replace a random
	// existing power.
	for i := 0; i < NumPowerSlots; i++ {
		if crit.Powers[i] == NoPower {
			crit.Powers[i] = power
			return
		}
	}
	crit.Powers[rand.Intn(NumPowerSlots)] = power
}

func cryoMutation(id entity.Id) {
	EMsg("{sub.Thename} feel{sub.s} very cold.\n", id, entity.NilId)
	addPower(id, PowerCryoBurst)
}

func lightningMutation(id entity.Id) {
	EMsg("{sub.Thename} feel{sub.s} very tingly.\n", id, entity.NilId)
	addPower(id, PowerChainLightning)
}
