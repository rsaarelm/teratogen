package main

import (
//	"hyades/entity"
)

type entityPrototype struct {
	Name     string
	Parent   string
	IconId   string
	Class    EntityClass
	Scarcity int
	MinDepth int
	Props    map[string]interface{}
}

// Keyword argument emulation with maps
type KW map[string]interface{}

func NewPrototype(name, parent, iconId string, class EntityClass, scarcity, minDepth int, kwargs KW) (result *entityPrototype) {
	result = new(entityPrototype)
	result.Name = name
	result.Parent = parent
	result.IconId = iconId
	result.Class = class
	result.Scarcity = scarcity
	result.MinDepth = minDepth

	result.Props = make(map[string]interface{})
	for k, v := range kwargs {
		result.Props[k] = v
	}
	return
}

func (self *entityPrototype) applyProps(prototypes map[string]*entityPrototype, target *Blob) {
	if parent, ok := prototypes[self.Parent]; ok {
		parent.applyProps(prototypes, target)
	}
	for key, val := range self.Props {
		target.Set(key, val)
	}
}

func (self *entityPrototype) MakeEntity(prototypes map[string]*entityPrototype, target *Blob) {
	target.IconId = self.IconId
	target.Name = self.Name
	target.Class = self.Class
	self.applyProps(prototypes, target)
}

var prototypes = map[string]*entityPrototype{
	// Base prototype for creatures.
	"creature": NewPrototype("creature", "", "", EnemyEntityClass, -1, 0, KW{
		FlagObstacle: 1,
		PropStrength: Fair,
		PropToughness: Fair,
		PropMeleeSkill: Fair,
		PropScale: 0,
		PropWounds: 0,
		PropDensity: 0}),
	"protagonist": NewPrototype("protagonist", "creature", "chars:0", PlayerEntityClass, -1, 0, KW{
		PropStrength: Great,
		PropToughness: Good,
		PropMeleeSkill: Good}),
	"zombie": NewPrototype("zombie", "creature", "chars:1", EnemyEntityClass, 100, 0, KW{
		PropStrength: Fair,
		PropToughness: Poor,
		PropMeleeSkill: Fair}),
	"dogthing": NewPrototype("dog-thing", "creature", "chars:2", EnemyEntityClass, 150, 0, KW{
		PropStrength: Fair,
		PropToughness: Fair,
		PropMeleeSkill: Good,
		PropScale: -1}),
	"ogre": NewPrototype("ogre", "creature", "chars:15", EnemyEntityClass, 600, 5, KW{
		PropStrength: Great,
		PropToughness: Great,
		PropMeleeSkill: Fair,
		PropScale: 3}),
	"boss1": NewPrototype("elder spawn", "creature", "chars:5", EnemyEntityClass, 3000, 10, KW{
		PropStrength: Legendary,
		PropToughness: Legendary,
		PropMeleeSkill: Superb,
		PropScale: 5}),
	"globe": NewPrototype("health globe", "", "items:1", GlobeEntityClass, 30, 0, KW{}),
	"plantpot": NewPrototype("plant pot", "", "items:3", ItemEntityClass, 200, 0, KW{}),
	"pistol": NewPrototype("pistol", "", "items:4", ItemEntityClass, 200, 0, KW{
		PropEquipmentSlot: PropGunWeaponGuid,
		PropWoundBonus: 1,
		PropDurability: 12}),
	"machete": NewPrototype("machete", "", "items:5", ItemEntityClass, 200, 0, KW{
		PropEquipmentSlot: PropMeleeWeaponGuid,
		PropWoundBonus: 2,
		PropDurability: 20}),
	"kevlar": NewPrototype("kevlar armor", "", "items:6", ItemEntityClass, 200, 0, KW{
		PropEquipmentSlot: PropBodyArmorGuid,
		PropToughness: Good,
		PropDefenseBonus: 1,
		PropDurability: 20}),
	"medkit": NewPrototype("medkit", "", "items:7", ItemEntityClass, 200, 0, KW{
		PropItemUse: MedkitUse}),
}
