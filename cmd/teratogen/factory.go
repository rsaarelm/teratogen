package main

import (
	"hyades/entity"
)

// Keyword argument emulation with maps
type KW map[string]interface{}


// The Metadata doesn't match any entity component, nor does it affect the
// entities created. It is for the use of the entity generator algorithm,
// which uses it to tell how likely the different entitities are to spawn.
const Metadata = entity.ComponentFamily("dummy-metadata")

type metaTemplate struct {
	Scarcity int
	MinDepth int
}

func MetaTemplate(scarcity, minDepth int) *metaTemplate {
	return &metaTemplate{scarcity, minDepth}
}

func (self *metaTemplate) Derive(c entity.ComponentTemplate) entity.ComponentTemplate {
	return c
}

func (self *metaTemplate) MakeComponent(manager *entity.Manager, guid entity.Id) {
	// no-op
}


type blobTemplate struct {
	IconId string
	Class  EntityClass
	Props  map[string]interface{}
}

func BlobTemplate(iconId string, class EntityClass, kwargs KW) (result *blobTemplate) {
	result = new(blobTemplate)
	result.IconId = iconId
	result.Class = class

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
	result := BlobTemplate(child.IconId, child.Class, KW{})
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
	blob.Class = self.Class
	self.applyProps(blob)
	blobs.Add(guid, blob)
}

var assemblages map[string]entity.Assemblage

func init() {
	assemblages = make(map[string]entity.Assemblage)
	a := assemblages
	a["creature"] = entity.Assemblage{
		Metadata: MetaTemplate(-1, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("creature"),
		BlobComponent: BlobTemplate("", EnemyEntityClass, KW{
			FlagObstacle: 1,
			PropStrength: Fair,
			PropToughness: Fair,
			PropMeleeSkill: Fair,
			PropScale: 0,
			PropDensity: 0})}
	a["protagonist"] = a["creature"].Derive(entity.Assemblage{
		Metadata: MetaTemplate(-1, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("protagonist"),
		CreatureComponent: &CreatureTemplate{
			Str: Great,
			Tough: Good,
			Melee: Good,
			Scale: 0,
			Density: 0},
		BlobComponent: BlobTemplate("chars:0", PlayerEntityClass, KW{
			PropStrength: Great,
			PropToughness: Good,
			PropMeleeSkill: Good})})
	a["zombie"] = a["creature"].Derive(entity.Assemblage{
		Metadata: MetaTemplate(100, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("zombie"),
		CreatureComponent: &CreatureTemplate{
			Str: Fair,
			Tough: Poor,
			Melee: Fair,
			Scale: 0,
			Density: 0},
		BlobComponent: BlobTemplate("chars:1", EnemyEntityClass, KW{
			PropStrength: Fair,
			PropToughness: Poor,
			PropMeleeSkill: Fair})})
	a["dogthing"] = a["creature"].Derive(entity.Assemblage{
		Metadata: MetaTemplate(150, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("dog-thing"),
		CreatureComponent: &CreatureTemplate{
			Str: Fair,
			Tough: Fair,
			Melee: Good,
			Scale: -1,
			Density: 0},
		BlobComponent: BlobTemplate("chars:2", EnemyEntityClass, KW{
			PropStrength: Fair,
			PropToughness: Fair,
			PropMeleeSkill: Good,
			PropScale: -1})})
	a["ogre"] = a["creature"].Derive(entity.Assemblage{
		Metadata: MetaTemplate(600, 5),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("ogre"),
		CreatureComponent: &CreatureTemplate{
			Str: Great,
			Tough: Great,
			Melee: Fair,
			Scale: 3,
			Density: 0},
		BlobComponent: BlobTemplate("chars:15", EnemyEntityClass, KW{
			PropStrength: Great,
			PropToughness: Great,
			PropMeleeSkill: Fair,
			PropScale: 3})})
	a["boss1"] = a["creature"].Derive(entity.Assemblage{
		Metadata: MetaTemplate(3000, 10),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("elder spawn"),
		CreatureComponent: &CreatureTemplate{
			Str: Legendary,
			Tough: Legendary,
			Melee: Superb,
			Scale: 5,
			Density: 0},
		BlobComponent: BlobTemplate("chars:5", EnemyEntityClass, KW{
			PropStrength: Legendary,
			PropToughness: Legendary,
			PropMeleeSkill: Superb,
			PropScale: 5})})

	a["globe"] = entity.Assemblage{
		Metadata: MetaTemplate(30, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("health globe"),
		BlobComponent: BlobTemplate("items:1", GlobeEntityClass, KW{})}
	a["plantpot"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("plant pot"),
		BlobComponent: BlobTemplate("items:3", ItemEntityClass, KW{})}
	a["pistol"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("pistol"),
		BlobComponent: BlobTemplate("items:4", ItemEntityClass, KW{
			PropEquipmentSlot: PropGunWeaponGuid,
			PropWoundBonus: 1,
			PropDurability: 12})}
	a["machete"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("machete"),
		BlobComponent: BlobTemplate("items:5", ItemEntityClass, KW{
			PropEquipmentSlot: PropMeleeWeaponGuid,
			PropWoundBonus: 2,
			PropDurability: 20})}
	a["kevlar"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("kevlar armor"),
		BlobComponent: BlobTemplate("items:6", ItemEntityClass, KW{
			PropEquipmentSlot: PropBodyArmorGuid,
			PropToughness: Good,
			PropDefenseBonus: 1,
			PropDurability: 20})}
	a["medkit"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("medkit"),
		BlobComponent: BlobTemplate("items:7", ItemEntityClass, KW{
			PropItemUse: MedkitUse})}
}
