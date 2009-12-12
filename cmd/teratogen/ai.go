package main

import . "hyades/gamelib"

func DoAI(crit *Creature) {
	world := GetWorld();
	player := world.GetPlayer();
	if player == nil || player == crit {
		return;
	}

	dirVec := player.GetPos().Minus(crit.GetPos());
	dir8 := Vec2IToDir8(dirVec);
	moveVec := Dir8ToVec(dir8);

	if crit.GetPos().Plus(moveVec).Equals(player.GetPos()) {
		Attack(crit, player);
	} else {
		// TODO: Going around obstacles.
		crit.TryMove(moveVec);
	}
}
