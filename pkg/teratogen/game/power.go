package game

import (
	"exp/iterable"
	"hyades/entity"
	"hyades/geom"
)

type PowerId int

const (
	NoPower PowerId = iota
	PowerCryoBurst
	PowerChainLightning
	PowerKineticBlast
	PowerTeleport
)

const NumPowerSlots = 4

type Power struct {
	ShortName string
	Cooldown  int

	// This must be a function value with one of the agreed-upon signatures for
	// power actions. The signature can differ in whether the power takes a
	// direction argument or not, and the power-using interface will want to
	// reflect this. The function will return a boolean which tells whether
	// using the power ended the turn for the power-user.
	Action interface{}
}

var powerLookup = map[PowerId]*Power{
	NoPower:        nil,
	PowerCryoBurst: &Power{"cryo burst", 5, DoCryoBurst},
}

func GetPower(id PowerId) *Power {
	return powerLookup[id]
}

func DoCryoBurst(user entity.Id, dir6 int) (endsMove bool) {
	var pos geom.Pt2I
	if dir6 == -1 {
		pos = GetPos(user)
	} else {
		dx, dy := geom.HexToPlane(geom.Origin.Plus(geom.Dir6ToVec(dir6)))

		// XXX: Shooting code should go into its own function.
	Outer:
		for o := range iterable.Drop(
			geom.HexRay(GetPos(user), dx, dy),
			1).Iter() {
			pos = o.(geom.Pt2I)
			if !IsOpenTerrain(pos) {
				break
			}

			for o := range EntitiesAt(pos).Iter() {
				id := o.(entity.Id)
				if BlocksMovement(id) {
					break Outer
				}
			}
		}
	}
	Fx().Shoot(user, pos, AttackFxFrost)

	// TODO: Make the explosion frost-colored.
	Fx().Explode(pos, 5, 2)

	for pt := range geom.HexRadiusIter(pos.X, pos.Y, 1) {
		critId := CreatureAt(pt)
		if crit := GetCreature(critId); crit != nil {
			// Let's be nice and not have friendly fire. You can't freeze
			// yourself with the blast.
			if critId != user {
				// XXX: No resistance check.
				crit.AddStatus(StatusFrozen)
				EMsg("{sub.Thename} {sub.is} frozen.\n", critId, entity.NilId)
			}
		}
	}

	// TODO: Freeze effect instead of damage.
	return true
}
