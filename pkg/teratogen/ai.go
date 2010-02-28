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

	dirVec := GetPos(playerId).Minus(GetPos(critId))
	dir6 := geom.Vec2IToDir6(dirVec)
	moveVec := geom.Dir6ToVec(dir6)

	crit := GetCreature(critId)

	// Bile attack.
	if crit.Traits&IntrinsicBile != 0 {
		if CanSeeTo(GetPos(critId), GetPos(playerId)) && num.OneChanceIn(2) {
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

func SendPlayerInput(command func()) bool {
	// Don't block, if the channel isn't expecting input, just move on and
	// return false.
	ok := playerInputChan <- command
	return ok
}

var playerInputChan = make(chan func())

// WaitPlayerInput is run in UI mode, it waits on the player input channel
// until some input arrives. Then it returns the thunk containing that input.
func WaitPlayerInput() func() { return <-playerInputChan }
