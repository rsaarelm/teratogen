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
	Props  map[string]interface{}
}

func BlobTemplate(iconId string, kwargs KW) (result *blobTemplate) {
	result = new(blobTemplate)
	result.IconId = iconId

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
	result := BlobTemplate(child.IconId, KW{})
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
	self.applyProps(blob)
	blobs.Add(guid, blob)
}

var assemblages map[string]entity.Assemblage

func init() {
	assemblages = make(map[string]entity.Assemblage)
	a := assemblages
	a["protagonist"] = entity.Assemblage{
		Metadata: MetaTemplate(-1, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("protagonist"),
		CreatureComponent: &CreatureTemplate{
			Str: Great,
			Tough: Good,
			Melee: Good,
			Scale: 0,
			Density: 0},
		BlobComponent: BlobTemplate("chars:0", KW{})}
	a["zombie"] = entity.Assemblage{
		Metadata: MetaTemplate(100, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("zombie"),
		CreatureComponent: &CreatureTemplate{
			Str: Fair,
			Tough: Poor,
			Melee: Fair,
			Scale: 0,
			Density: 0},
		BlobComponent: BlobTemplate("chars:1", KW{})}
	a["dogthing"] = entity.Assemblage{
		Metadata: MetaTemplate(150, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("dog-thing"),
		CreatureComponent: &CreatureTemplate{
			Str: Fair,
			Tough: Fair,
			Melee: Good,
			Scale: -1,
			Density: 0},
		BlobComponent: BlobTemplate("chars:2", KW{})}
	a["ogre"] = entity.Assemblage{
		Metadata: MetaTemplate(600, 5),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("ogre"),
		CreatureComponent: &CreatureTemplate{
			Str: Great,
			Tough: Great,
			Melee: Fair,
			Scale: 3,
			Density: 0},
		BlobComponent: BlobTemplate("chars:15", KW{})}
	a["boss1"] = entity.Assemblage{
		Metadata: MetaTemplate(3000, 10),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("elder spawn"),
		CreatureComponent: &CreatureTemplate{
			Str: Legendary,
			Tough: Legendary,
			Melee: Superb,
			Scale: 5,
			Density: 0},
		BlobComponent: BlobTemplate("chars:5", KW{})}

	a["globe"] = entity.Assemblage{
		Metadata: MetaTemplate(30, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("health globe"),
		BlobComponent: BlobTemplate("items:1", KW{})}
	a["plantpot"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("plant pot"),
		ItemComponent: &ItemTemplate{NoEquipSlot, 0, 0, 0, NoUse},
		BlobComponent: BlobTemplate("items:3", KW{})}
	a["pistol"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("pistol"),
		ItemComponent: &ItemTemplate{
			EquipmentSlot: GunEquipSlot,
			Durability: 12,
			WoundBonus: 1,
			DefenseBonus: 0,
			Use: NoUse},
		BlobComponent: BlobTemplate("items:4", KW{})}
	a["machete"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("machete"),
		ItemComponent: &ItemTemplate{
			EquipmentSlot: MeleeEquipSlot,
			Durability: 20,
			WoundBonus: 2,
			DefenseBonus: 0,
			Use: NoUse},
		BlobComponent: BlobTemplate("items:5", KW{})}
	a["kevlar"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("kevlar armor"),
		ItemComponent: &ItemTemplate{
			EquipmentSlot: ArmorEquipSlot,
			Durability: 20,
			WoundBonus: 0,
			DefenseBonus: 1,
			Use: NoUse},
		BlobComponent: BlobTemplate("items:6", KW{})}
	a["medkit"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate("medkit"),
		ItemComponent: &ItemTemplate{NoEquipSlot, 0, 0, 0, MedkitUse},
		BlobComponent: BlobTemplate("items:7", KW{})}
}
