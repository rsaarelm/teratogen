package main

import (
	"hyades/geom"
)

func DoAI(crit *Entity) {
	world := GetWorld()
	player := world.GetPlayer()
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
