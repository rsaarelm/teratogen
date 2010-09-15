package teratogen

import (
	"exp/iterable"
	"hyades/alg"
	"hyades/geom"
	"hyades/entity"
	"hyades/num"
	"rand"
)

// DoAI runs AI for a creature. Returns true if this ends the creature's move,
// false if the creature gets to act again immediately afterwards.
func DoAI(aiEntity entity.Id) bool {
	playerId := PlayerId()
	if aiEntity == playerId {
		return true
	}

	crit := GetCreature(aiEntity)
	if crit == nil {
		// It's probably been killed mid-iteration.
		return true
	}

	if !CanAct(aiEntity) {
		return true
	}

	// Occasional secondary attacks.
	if num.OneChanceIn(2) && aiWeaponAttack(aiEntity, weaponLookup[crit.Attack2]) {
		return true
	}

	// Otherwise try the primary attack always.
	if aiWeaponAttack(aiEntity, weaponLookup[crit.Attack1]) {
		return true
	}

	aiMove(aiEntity)

	return true
}

func aiMove(aiEntity entity.Id) {
	dirVec := GetPos(aiGetCurrentEnemyId(aiEntity)).Minus(GetPos(aiEntity))
	dir6 := geom.Vec2IToDir6(dirVec)
	moveVec := geom.Dir6ToVec(dir6)

	if TryMove(aiEntity, moveVec) {
		return
	}
	// Random move if directional move is blocked.
	TryMove(aiEntity, geom.Dir6ToVec(rand.Intn(6)))
}

// aiWeaponAttack tries to attack an AI-decided target with a given weapon.
// Returns whether the attack succeeded. Weapon may be nil, in which case the
// function does nothing and returns false.
func aiWeaponAttack(aiEntity entity.Id, weapon *Weapon) bool {
	if weapon == nil {
		return false
	}

	enemyId := aiGetCurrentEnemyId(aiEntity)
	if enemyId == entity.NilId {
		return false
	}
	pos := GetPos(enemyId)
	if CanSeeTo(GetPos(aiEntity), pos) && weapon.CanAttack(aiEntity, pos) {
		// XXX: Replicating AttackSkill check. Put attack logic in one place only.
		weapon.Attack(aiEntity, pos, AttackSkill(aiEntity, weapon))
		return true
	}
	return false
}

// aiGetCurrentEnemy returns the entity which the AI entity is currently
// attacking. May return entity.NilId if there is no current enemy.
func aiGetCurrentEnemyId(aiEntity entity.Id) (currentEnemy entity.Id) {
	// Currently everything just hunts the player.
	return PlayerId()
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
