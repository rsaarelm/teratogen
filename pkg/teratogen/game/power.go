package game

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

func ExpectsDirection(power PowerId) bool {
	switch power {
	case PowerCryoBurst,
		PowerKineticBlast,
		PowerTeleport:
		return true
	}
	return false
}

func ShortPowerName(power PowerId) string {
	switch power {
	//        "--------------" Size limit
	case PowerCryoBurst:
		return "cryo burst"
	case PowerChainLightning:
		return "chain lightng"
	case PowerKineticBlast:
		return "kinetic blast"
	case PowerTeleport:
		return "teleport"
	}
	return "UNKNOWN POWER"
}

func UsePower(user entity.Id, power PowerId, dir6 int) {
	switch power {
	// TODO
	}
}
