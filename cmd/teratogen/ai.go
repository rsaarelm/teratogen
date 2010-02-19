package main

import (
	"hyades/geom"
)

func DoAI(crit *Blob) {
	player := GetPlayer()
	if player == nil || player == crit {
		return
	}

	dirVec := player.GetPos().Minus(crit.GetPos())
	dir8 := geom.Vec2IToDir8(dirVec)
	moveVec := geom.Dir8ToVec(dir8)

	if crit.GetPos().Plus(moveVec).Equals(player.GetPos()) {
		Attack(crit, player)
	} else {
		// TODO: Going around obstacles.
		crit.TryMove(moveVec)
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
