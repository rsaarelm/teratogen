package teratogen

import (
	"hyades/entity"
	"hyades/num"
)

// DoTurn is the entry point for running the game update loop.
func DoTurn() {
	makeVelocitiesStale()

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

	clearStaleVelocities()

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
func Heartbeat(id entity.Id) {
	bloodtrailHeartbeat(id)
	buffHeartbeat(id)
}

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

func buffHeartbeat(id entity.Id) {
	crit := GetCreature(id)
	if crit == nil {
		return
	}

	if crit.SaveToLose(StatusConfused, 1.0/10) {
		EMsg("{sub.Thename} {sub.is} no longer %s.\n",
			id, entity.NilId, StatusDescription(StatusConfused))
	}
}

// Make the velocities of positional entities stale, call this at the
// beginning of a turn.
func makeVelocitiesStale() {
	for o := range Creatures().Iter() {
		id := o.(entity.Id)
		if pos := PosComp(id); EntityActsThisTurn(id) && pos != nil {
			pos.MakeVelocityStale()
		}
	}
}

// Clear velocities that are stale. Call this at the end of the turn, to set
// velocities of unmoved entities to zero.
func clearStaleVelocities() {
	for o := range Creatures().Iter() {
		id := o.(entity.Id)
		if pos := PosComp(id); EntityActsThisTurn(id) && pos != nil {
			pos.ClearStaleVelocity()
		}
	}
}
