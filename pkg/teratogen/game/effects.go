package game

import (
	"hyades/entity"
	"hyades/geom"
)

type AttackFx int

const (
	NoAttackFx AttackFx = iota
	AttackFxBeam
	AttackFxSpray
	AttackFxElectro
)

// Interface which Teratogen uses to communicate events to the client module.
type Effects interface {
	// Print prints some text to the client message panel.
	Print(str string)

	// Shoot shows an entity shooting somewhere.
	Shoot(shooterId entity.Id, target geom.Pt2I, fx AttackFx)

	// Damage shows damage done to an entity.
	Damage(id entity.Id, amout int)

	// Sparks shows a spark blast at the given pos. Used for example for
	// bullets hitting walls. This might be changed into a more general "show
	// effect here" method later.
	Sparks(pos geom.Pt2I)

	// Heal shows an entity healing.
	Heal(id entity.Id, amount int)

	// Destroy shows an entity being destroyed
	Destroy(id entity.Id)

	// Quit signals that the game is over.
	Quit(message string)

	// MorePrompt tells the player to pay attention to something important.
	MorePrompt()

	// Show an explosion around center
	Explode(center geom.Pt2I, power int, radius int)

	// Wait for player input from UI. The return value func returns true if the
	// action ends the player's move and false if the player can perform
	// another move.
	GetPlayerInput() func() bool
}
