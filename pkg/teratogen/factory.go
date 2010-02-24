package teratogen

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


var assemblages map[string]entity.Assemblage

func init() {
	assemblages = make(map[string]entity.Assemblage)
	a := assemblages
	a["protagonist"] = entity.Assemblage{
		Metadata: MetaTemplate(-1, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"protagonist", "chars:0"},
		CreatureComponent: &CreatureTemplate{
			Str: Great,
			Tough: Good,
			Melee: Good,
			Scale: 0,
			Density: 0,
			Traits: NoIntrinsic},
	}
	a["zombie"] = entity.Assemblage{
		Metadata: MetaTemplate(100, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"zombie", "chars:1"},
		CreatureComponent: &CreatureTemplate{
			Str: Fair,
			Tough: Poor,
			Melee: Fair,
			Scale: 0,
			Density: 0,
			Traits: NoIntrinsic},
	}
	a["dogthing"] = entity.Assemblage{
		Metadata: MetaTemplate(150, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"dog-thing", "chars:2"},
		CreatureComponent: &CreatureTemplate{
			Str: Fair,
			Tough: Fair,
			Melee: Good,
			Scale: -1,
			Density: 0,
			Traits: NoIntrinsic},
	}
	a["belcher"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 2),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"belcher", "chars:3"},
		CreatureComponent: &CreatureTemplate{
			Str: Poor,
			Tough: Mediocre,
			Melee: Mediocre,
			Scale: 1,
			Density: 0,
			Traits: IntrinsicBile},
	}
	a["crawlingmass"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 4),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"crawling mass", "chars:7"},
		CreatureComponent: &CreatureTemplate{
			Str: Good,
			Tough: Good,
			Melee: Poor,
			Scale: 2,
			Density: 0,
			Traits: IntrinsicSlow | IntrinsicDeathsplode},
	}
	a["cyclops"] = entity.Assemblage{
		Metadata: MetaTemplate(300, 4),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"cyclops", "chars:6"},
		CreatureComponent: &CreatureTemplate{
			Str: Poor,
			Tough: Good,
			Melee: Poor,
			Scale: 0,
			Density: 0,
			Traits: IntrinsicPsychicBlast | IntrinsicConfuse},
	}
	a["wendigo"] = entity.Assemblage{
		Metadata: MetaTemplate(300, 6),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"wendigo", "chars:8"},
		CreatureComponent: &CreatureTemplate{
			Str: Superb,
			Tough: Fair,
			Melee: Superb,
			Scale: 2,
			Density: 0,
			Traits: IntrinsicFast},
	}
	a["spider"] = entity.Assemblage{
		Metadata: MetaTemplate(400, 6),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"spider", "chars:10"},
		CreatureComponent: &CreatureTemplate{
			Str: Fair,
			Tough: Fair,
			Melee: Great,
			Scale: 1,
			Density: 0,
			Traits: IntrinsicPoison},
	}
	a["killbot"] = entity.Assemblage{
		Metadata: MetaTemplate(300, 8),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"killbot", "chars:11"},
		CreatureComponent: &CreatureTemplate{
			Str: Superb,
			Tough: Great,
			Melee: Good,
			Scale: 0,
			Density: 2,
			Traits: IntrinsicElectrocute},
	}
	a["ogre"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 8),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"ogre", "chars:15"},
		CreatureComponent: &CreatureTemplate{
			Str: Great,
			Tough: Great,
			Melee: Fair,
			Scale: 3,
			Density: 0,
			Traits: NoIntrinsic},
	}
	a["boss1"] = entity.Assemblage{
		Metadata: MetaTemplate(3000, 12),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"elder spawn", "chars:5"},
		CreatureComponent: &CreatureTemplate{
			Str: Legendary,
			Tough: Legendary,
			Melee: Superb,
			Scale: 5,
			Density: 0,
			Traits: NoIntrinsic},
	}

	a["globe"] = entity.Assemblage{
		Metadata: MetaTemplate(30, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"health globe", "items:1"},
	}
	a["plantpot"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"plant pot", "items:3"},
		ItemComponent: &ItemTemplate{NoEquipSlot, 0, 0, 0, NoUse},
	}
	a["pistol"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"pistol", "items:4"},
		ItemComponent: &ItemTemplate{
			EquipmentSlot: GunEquipSlot,
			Durability: 12,
			WoundBonus: 1,
			DefenseBonus: 0,
			Use: NoUse},
	}
	a["machete"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"machete", "items:5"},
		ItemComponent: &ItemTemplate{
			EquipmentSlot: MeleeEquipSlot,
			Durability: 20,
			WoundBonus: 2,
			DefenseBonus: 0,
			Use: NoUse},
	}
	a["kevlar"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"kevlar armor", "items:6"},
		ItemComponent: &ItemTemplate{
			EquipmentSlot: ArmorEquipSlot,
			Durability: 20,
			WoundBonus: 0,
			DefenseBonus: 1,
			Use: NoUse},
	}
	a["medkit"] = entity.Assemblage{
		Metadata: MetaTemplate(200, 0),
		PosComponent: PosTemplate(),
		NameComponent: NameTemplate{"medkit", "items:7"},
		ItemComponent: &ItemTemplate{NoEquipSlot, 0, 0, 0, MedkitUse},
	}
}
