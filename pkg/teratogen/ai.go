package teratogen

import (
	"exp/iterable"
	"hyades/alg"
	"hyades/geom"
	"hyades/entity"
	"hyades/num"
)

func DoAI(critId entity.Id) bool {
	playerId := PlayerId()
	if critId == playerId {
		return true
	}

	crit := GetCreature(critId)
	if crit == nil {
		// It's probably been killed mid-iteration.
		return true
	}

	dirVec := GetPos(playerId).Minus(GetPos(critId))
	dir6 := geom.Vec2IToDir6(dirVec)
	moveVec := geom.Dir6ToVec(dir6)

	// Bile attack.
	const bileAttackRange = 5
	if crit.HasIntrinsic(IntrinsicBile) && num.OneChanceIn(2) {
		if targetPos, ok := BestAttackTarget(critId, 1, bileAttackRange); ok {
			damageFactor := crit.Power + crit.Scale

			Fx().Shoot(critId, targetPos)
			EMsg("{sub.Thename} vomit{sub.s} bile.\n", critId, entity.NilId)
			DamagePos(targetPos, GetPos(critId), float64(damageFactor), AcidDamage, critId)

			return true
		}
	}

	if GetPos(critId).Plus(moveVec).Equals(GetPos(playerId)) {
		Attack(critId, playerId)
	} else {
		// TODO: Going around obstacles.
		TryMove(critId, moveVec)
	}
	return true
}

// DoTurn is the entry point for running the game update loop.
func DoTurn() {
	if IsHeartbeatTurn(GetCurrentTurn()) {
		Heartbeats()
	}

	levelNum := GetCurrentLevel()

	if PlayerActsThisTurn() {
		for {
			playerInput := Fx().GetPlayerInput()
			endPlayerMove := playerInput()
			if endPlayerMove {
				break
			}
		}
	}

	if GetCurrentLevel() != levelNum {
		// XXX: Hack: If the player entered a new level, skip the rest of the
		// turn, don't let the creatures on the next level act before the
		// player.
		return
	}

	for o := range Creatures().Iter() {
		id := o.(entity.Id)
		if id == PlayerId() || !EntityActsThisTurn(id) {
			continue
		}
		for {
			endCreatureMove := DoAI(id)
			if endCreatureMove {
				break
			}
		}
	}

	NextTurn()
}

// ActsOnTurn returns whether an active entity can act on a given turn
// based on its speed intrinsics. The system was originally described by Jeff
// Lait, Message-ID: <774acfb8.0410200424.1d09a21e@posting.google.com>
func ActsOnTurn(turnNum int64, isSlow, isFast, isQuick bool) bool {
	switch turnNum % 5 {
	case 0:
		// Fast phase, only fast entities act.
		return isFast
	case 1:
		// Normal phase, all entities act.
		return true
	case 2:
		// Slow phase, only non-slow entities act.
		return !isSlow
	case 3:
		// Quick phase, only quick entities act.
		return isQuick
	case 4:
		// Normal phase, all entities act.
		return true
	}
	// Shouldn't end up here...
	return false
}

func EntityActsThisTurn(id entity.Id) bool {
	crit := GetCreature(id)
	if crit == nil {
		return false
	}

	turn := GetCurrentTurn()
	isSlow := crit.Intrinsics&IntrinsicSlow != 0
	isFast := crit.Intrinsics&IntrinsicFast != 0
	isQuick := crit.Statuses&StatusQuick != 0

	return ActsOnTurn(turn, isSlow, isFast, isQuick)
}

// IsHeartBeatTurn returns whether a the heartbeat entity update should be run
// on the given turn. Based on the same system as EntityActsOnTurn. Heartbeat
// occurs on Normal and Slow phases, but not in the fast or quick phases.
func IsHeartbeatTurn(turnNum int64) bool {
	switch turnNum % 5 {
	case 1, 2, 4:
		return true
	}
	return false
}

func PlayerActsThisTurn() bool { return EntityActsThisTurn(PlayerId()) }

func Heartbeats() {
	for o := range Creatures().Iter() {
		Heartbeat(o.(entity.Id))
	}
}

// Heartbeat runs status updates on active entities. Things such as temporary
// effects wearing off or affecting an entity go here.
func Heartbeat(id entity.Id) { bloodtrailHeartbeat(id) }

func bloodtrailHeartbeat(id entity.Id) {
	crit := GetCreature(id)
	if crit == nil {
		return
	}

	standingIn := BloodSplatterAt(GetPos(id))
	if standingIn == LargeBloodSplatter {
		// Creatures start tracking blood when they walk through pools of blood.
		crit.AddStatus(StatusBloodTrail)
	} else {
		if crit.HasStatus(StatusBloodTrail) {
			SplatterBlood(GetPos(id), BloodTrail)
			if num.OneChanceIn(3) {
				crit.RemoveStatus(StatusBloodTrail)
			}
		}
	}
}

// Return whether an entity considers another entity an enemy.
func IsEnemyOf(id, possibleEnemyId entity.Id) bool {
	if id == possibleEnemyId {
		return false
	}

	if id == entity.NilId || possibleEnemyId == entity.NilId {
		return false
	}

	// XXX: Currently player is the enemy of every other creature. This should
	// be replaced with a more general faction system.
	if IsCreature(id) && IsCreature(possibleEnemyId) &&
		(id == PlayerId() || possibleEnemyId == PlayerId()) {
		return true
	}

	return false
}

// EnemiesAt iterates the enemies of ent at pos.
func EnemiesAt(id entity.Id, pos geom.Pt2I) iterable.Iterable {
	filter := func(o interface{}) bool { return IsEnemyOf(id, o.(entity.Id)) }

	return iterable.Filter(EntitiesAt(pos), filter)
}

func CreaturesSeenBy(o interface{}) iterable.Iterable {
	id := o.(entity.Id)
	pred := func(o interface{}) bool { return CanSeeTo(GetPos(id), GetPos(o.(entity.Id))) }
	return iterable.Filter(OtherCreatures(o), pred)
}

func ClosestCreatureSeenBy(id entity.Id) entity.Id {
	distFromSelf := func(idOther interface{}) float64 { return EntityDist(idOther.(entity.Id), id) }
	ret, ok := alg.IterMin(CreaturesSeenBy(id), distFromSelf)
	if !ok {
		return entity.NilId
	}
	return ret.(entity.Id)
}

// AttackTargetPredicate returns a function that tells whether there's
// something in the given position that the entity attacker wants to hurt. The
// result function expects geom.Pt2I, but has interface{} type to make it
// simpler to use with with iterable.Filter.
func AttackTargetPredicate(attacker entity.Id) func(interface{}) bool {
	return func(pos interface{}) bool {
		return !alg.IsEmptyIter(EnemiesAt(attacker, pos.(geom.Pt2I)))
	}
}

// AttackPriority returns a function that gives the priority with which
// a creature wants to attack the given position. Lower value indicates higher
// priority.
func AttackPriority(attacker entity.Id) func(interface{}) float64 {
	return func(o interface{}) float64 {
		pos := o.(geom.Pt2I)
		// Add some noise to the results so that tie-breaks are less
		// predictable.
		return float64(geom.HexDist(GetPos(attacker), pos)) +
			num.SmoothNoise3D(float64(pos.X), float64(pos.Y), 0.0)/2.0
	}
}

// BestAttackTarget returns the most preferred target for attacker within the
// given range.
func BestAttackTarget(attacker entity.Id, minRange, maxRange int) (target geom.Pt2I, targetFound bool) {
	// Where the rays shot in 6 directions would stop.
	targets := geom.RayEndsIn6Dirs(
		GetPos(attacker),
		func(pt geom.Pt2I) bool { return !IsOpen(pt) },
		minRange,
		maxRange)
	// The ones that are something the attacker wants to hit.
	targets = iterable.Filter(targets, AttackTargetPredicate(attacker))
	o, targetFound := alg.IterMin(targets, AttackPriority(attacker))
	if targetFound {
		target = o.(geom.Pt2I)
	}
	return
}
