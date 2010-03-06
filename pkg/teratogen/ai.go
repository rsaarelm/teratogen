package teratogen

import (
	"hyades/geom"
	"hyades/entity"
	"hyades/num"
)

func DoAI(critId entity.Id) {
	playerId := PlayerId()
	if critId == playerId {
		return
	}

	crit := GetCreature(critId)
	if crit == nil {
		// It's probably been killed mid-iteration.
		return
	}

	dirVec := GetPos(playerId).Minus(GetPos(critId))
	dir8 := geom.Vec2IToDir8(dirVec)
	moveVec := geom.Dir8ToVec(dir8)

	// Bile attack.
	const bileAttackRange = 5
	if crit.Traits&IntrinsicBile != 0 {
		if CanSeeTo(GetPos(critId), GetPos(playerId)) && EntityDist(critId, playerId) <= bileAttackRange && num.OneChanceIn(2) {
			damageFactor := num.Imax(1, crit.Str+crit.Scale)

			hitPos := GetHitPos(GetPos(critId), GetPos(playerId))
			Fx().Shoot(critId, hitPos)
			Msg("%s vomits bile\n", GetCapName(critId))
			DamagePos(hitPos, damageFactor, critId)

			return
		}
	}

	if GetPos(critId).Plus(moveVec).Equals(GetPos(playerId)) {
		Attack(critId, playerId)
	} else {
		// TODO: Going around obstacles.
		TryMove(critId, moveVec)
	}
}

// DoTurn is the entry point for running the game update loop.
func DoTurn() {
	if IsHeartbeatTurn(GetCurrentTurn()) {
		Heartbeats()
	}

	if PlayerActsThisTurn() {
		playerInput := Fx().GetPlayerInput()
		playerInput()
	}

	for o := range Creatures().Iter() {
		id := o.(entity.Id)
		if id == PlayerId() || !EntityActsThisTurn(id) {
			continue
		}
		DoAI(id)
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
	isSlow := crit.Traits&IntrinsicSlow != 0
	isFast := crit.Traits&IntrinsicFast != 0
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
