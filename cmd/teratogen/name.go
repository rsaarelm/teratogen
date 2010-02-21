package main

import (
	"hyades/entity"
	"hyades/txt"
)

const NameComponent = entity.ComponentFamily("name")


type NameTemplate string

func (self NameTemplate) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return c
}

func (self NameTemplate) MakeComponent(manager *entity.Manager, guid entity.Id) {
	GetManager().Handler(NameComponent).Add(guid, &Name{string(self)})
}


// Name component. Will probably get more structure than this eventually.
type Name struct {
	name string
}


// GetName returns the name of an entity. If the entity has no name component,
// it returns a string representation of its id value.
func GetName(id entity.Id) string {
	if nameComp := GetManager().Handler(NameComponent).Get(id); nameComp != nil {
		return nameComp.(*Name).name
	}
	return string(id)
}

// GetCapName returns the capitalized name of an entity.
func GetCapName(id entity.Id) string { return txt.Capitalize(GetName(id)) }
