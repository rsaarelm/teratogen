package teratogen

import . "fomalhaut"

func (self *World) DoAI(crit *Creature) {
	player := self.GetPlayer();
	if player == nil || player == crit {
		return;
	}

	dirVec := player.GetPos().Minus(crit.GetPos());
	dir8 := Vec2IToDir8(dirVec);
	moveVec := Dir8ToVec(dir8);

	if crit.GetPos().Plus(moveVec).Equals(player.GetPos()) {
		self.Attack(crit, player);
	} else {
		// TODO: Going around obstacles.
		self.MoveCreature(crit, moveVec);
	}
}