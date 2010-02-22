package main

import (
	"hyades/entity"
	"hyades/geom"
)

// Interface which Teratogen uses to communicate events to the client module.
type Effects interface {
	// Print prints some text to the client message panel.
	Print(str string)

	// Shoot shows an entity shooting somewhere.
	Shoot(shooterId entity.Id, target geom.Pt2I)

	// Damage shows damage done to an entity.
	Damage(id entity.Id, amout int)

	// Heal shows an entity healing.
	Heal(id entity.Id, amount int)

	// Destroy shows an entity being destroyed
	Destroy(id entity.Id)
}
