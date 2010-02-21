package main

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

func LogicLoop() {
	for {
		playerInput := <-playerInputChan
		MarkMsgLinesSeen()

		GetUISync()
		playerInput()
		RunAI()
		ReleaseUISync()
	}
}
