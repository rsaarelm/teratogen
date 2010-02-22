package teratogen

import (
	"hyades/entity"
	"hyades/txt"
)

const NameComponent = entity.ComponentFamily("name")


type NameTemplate Name

func (self NameTemplate) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return c
}

func (self NameTemplate) MakeComponent(manager *entity.Manager, guid entity.Id) {
	manager.Handler(NameComponent).Add(guid, &Name{self.Name, self.IconId})
}


// Name component. Will probably get more structure than this eventually.
type Name struct {
	Name   string
	IconId string
}


// GetName returns the name of an entity. If the entity has no name component,
// it returns a string representation of its id value.
func GetName(id entity.Id) string {
	if nameComp := GetManager().Handler(NameComponent).Get(id); nameComp != nil {
		return nameComp.(*Name).Name
	}
	return string(id)
}

func GetIconId(id entity.Id) string {
	if nameComp := GetManager().Handler(NameComponent).Get(id); nameComp != nil {
		return nameComp.(*Name).IconId
	}
	return ""
}

// GetCapName returns the capitalized name of an entity.
func GetCapName(id entity.Id) string { return txt.Capitalize(GetName(id)) }
