package main

import (
	"hyades/entity"
)

type blobTemplate struct {
	Name     string
	IconId   string
	Class    EntityClass
	Scarcity int
	MinDepth int
	Props    map[string]interface{}
}

// Keyword argument emulation with maps
type KW map[string]interface{}

func BlobTemplate(name, iconId string, class EntityClass, scarcity, minDepth int, kwargs KW) (result *blobTemplate) {
	result = new(blobTemplate)
	result.Name = name
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

func (self *blobTemplate) applyProps(target *Blob) {
	for key, val := range self.Props {
		target.Set(key, val)
	}
}

func (self *blobTemplate) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	child := c.(*blobTemplate)
	result := BlobTemplate(child.Name, child.IconId, child.Class,
		child.Scarcity, child.MinDepth, KW{})
	for key, val := range self.Props {
		result.Props[key] = val
	}

	for key, val := range child.Props {
		result.Props[key] = val
	}

	return result
}

func (self *blobTemplate) MakeComponent(manager *entity.Manager, guid entity.Id) {
	blobs := manager.Handler(BlobComponent)
	blob := NewEntity(guid)
	blob.IconId = self.IconId
	blob.Name = self.Name
	blob.Class = self.Class
	self.applyProps(blob)
	blobs.Add(guid, blob)
}

var assemblages map[string]entity.Assemblage

func init() {
	assemblages = make(map[string]entity.Assemblage)
	a := assemblages
	a["creature"] = entity.Assemblage{
		BlobComponent: BlobTemplate("creature", "", EnemyEntityClass, -1, 0, KW{
			FlagObstacle: 1,
			PropStrength: Fair,
			PropToughness: Fair,
			PropMeleeSkill: Fair,
			PropScale: 0,
			PropWounds: 0,
			PropDensity: 0})}
	a["protagonist"] = a["creature"].Derive(entity.Assemblage{
		BlobComponent: BlobTemplate("protagonist", "chars:0", PlayerEntityClass, -1, 0, KW{
			PropStrength: Great,
			PropToughness: Good,
			PropMeleeSkill: Good})})
	a["zombie"] = a["creature"].Derive(entity.Assemblage{
		BlobComponent: BlobTemplate("zombie", "chars:1", EnemyEntityClass, 100, 0, KW{
			PropStrength: Fair,
			PropToughness: Poor,
			PropMeleeSkill: Fair})})
	a["dogthing"] = a["creature"].Derive(entity.Assemblage{
		BlobComponent: BlobTemplate("dog-thing", "chars:2", EnemyEntityClass, 150, 0, KW{
			PropStrength: Fair,
			PropToughness: Fair,
			PropMeleeSkill: Good,
			PropScale: -1})})
	a["ogre"] = a["creature"].Derive(entity.Assemblage{
		BlobComponent: BlobTemplate("ogre", "chars:15", EnemyEntityClass, 600, 5, KW{
			PropStrength: Great,
			PropToughness: Great,
			PropMeleeSkill: Fair,
			PropScale: 3})})
	a["boss1"] = a["creature"].Derive(entity.Assemblage{
		BlobComponent: BlobTemplate("elder spawn", "chars:5", EnemyEntityClass, 3000, 10, KW{
			PropStrength: Legendary,
			PropToughness: Legendary,
			PropMeleeSkill: Superb,
			PropScale: 5})})

	a["globe"] = entity.Assemblage{
		BlobComponent: BlobTemplate("health globe", "items:1", GlobeEntityClass, 30, 0, KW{})}
	a["plantpot"] = entity.Assemblage{
		BlobComponent: BlobTemplate("plant pot", "items:3", ItemEntityClass, 200, 0, KW{})}
	a["pistol"] = entity.Assemblage{
		BlobComponent: BlobTemplate("pistol", "items:4", ItemEntityClass, 200, 0, KW{
			PropEquipmentSlot: PropGunWeaponGuid,
			PropWoundBonus: 1,
			PropDurability: 12})}
	a["machete"] = entity.Assemblage{
		BlobComponent: BlobTemplate("machete", "items:5", ItemEntityClass, 200, 0, KW{
			PropEquipmentSlot: PropMeleeWeaponGuid,
			PropWoundBonus: 2,
			PropDurability: 20})}
	a["kevlar"] = entity.Assemblage{
		BlobComponent: BlobTemplate("kevlar armor", "items:6", ItemEntityClass, 200, 0, KW{
			PropEquipmentSlot: PropBodyArmorGuid,
			PropToughness: Good,
			PropDefenseBonus: 1,
			PropDurability: 20})}
	a["medkit"] = entity.Assemblage{
		BlobComponent: BlobTemplate("medkit", "items:7", ItemEntityClass, 200, 0, KW{
			PropItemUse: MedkitUse})}
}
