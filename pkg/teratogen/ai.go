package teratogen

import (
	"hyades/geom"
	"hyades/entity"
)

func DoAI(critId entity.Id) {
	playerId := PlayerId()
	if critId == playerId {
		return
	}

	dirVec := GetPos(playerId).Minus(GetPos(critId))
	dir8 := geom.Vec2IToDir8(dirVec)
	moveVec := geom.Dir8ToVec(dir8)

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
func WaitPlayerInput() (func()) { return <-playerInputChan }
