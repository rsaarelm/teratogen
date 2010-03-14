package teratogen

import (
	"hyades/entity"
)

const DecalComponent = entity.ComponentFamily("decal")


type DecalTemplate int

func (self DecalTemplate) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return c
}

func (self DecalTemplate) MakeComponent(manager *entity.Manager, guid entity.Id) {
	manager.Handler(DecalComponent).Add(guid, &Decal{int(self)})
}

type Decal struct {
	DrawLayer int
}

func IsDecal(id entity.Id) bool { return GetManager().Handler(DecalComponent).Get(id) != nil }
