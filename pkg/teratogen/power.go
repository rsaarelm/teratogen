package teratogen

import (
	"hyades/entity"
)

type PowerId int

const (
	NoPower PowerId = iota
	PowerCryoBurst
	PowerChainLightning
	PowerKineticBlast
	PowerTeleport
)

const NumPowerSlots = 4

// Powers that don't need extra parameters to activate.
type UndirectedPower interface {
	UsePower(user entity.Id)
}

// Power that goes in a specific direction. Might take -1 for "point to self".
type DirectedPower interface {
	UsePowerTowards(user entity.Id, dir6 int)
}

// Power that targets a specific entity. Item enchantments and such.
type TargetedPower interface {
	UsePowerAt(user entity.Id, target entity.Id)
}
