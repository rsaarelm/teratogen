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
		Metadata:      MetaTemplate(-1, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"protagonist", "chars:0"},
		CreatureComponent: &CreatureTemplate{
			Power:  Great,
			Skill:  Good,
			Scale:  0,
			Traits: NoIntrinsic},
	}
	a["zombie"] = entity.Assemblage{
		Metadata:      MetaTemplate(100, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"zombie", "chars:1"},
		CreatureComponent: &CreatureTemplate{
			Power:  Fair,
			Skill:  Fair,
			Scale:  0,
			Traits: IntrinsicFragile},
	}
	a["dogthing"] = entity.Assemblage{
		Metadata:      MetaTemplate(150, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"dog-thing", "chars:2"},
		CreatureComponent: &CreatureTemplate{
			Power:  Fair,
			Skill:  Good,
			Scale:  -1,
			Traits: NoIntrinsic},
	}
	a["belcher"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 2),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"belcher", "chars:3"},
		CreatureComponent: &CreatureTemplate{
			Power:  Poor,
			Skill:  Mediocre,
			Scale:  1,
			Traits: IntrinsicBile | IntrinsicTough},
	}
	a["crawlingmass"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 4),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"crawling mass", "chars:7"},
		CreatureComponent: &CreatureTemplate{
			Power:  Good,
			Skill:  Poor,
			Scale:  2,
			Traits: IntrinsicSlow | IntrinsicDeathsplode},
	}
	a["cyclops"] = entity.Assemblage{
		Metadata:      MetaTemplate(300, 4),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"cyclops", "chars:6"},
		CreatureComponent: &CreatureTemplate{
			Power:  Poor,
			Skill:  Poor,
			Scale:  0,
			Traits: IntrinsicPsychicBlast | IntrinsicConfuse | IntrinsicTough},
	}
	a["wendigo"] = entity.Assemblage{
		Metadata:      MetaTemplate(300, 6),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"wendigo", "chars:8"},
		CreatureComponent: &CreatureTemplate{
			Power:  Superb,
			Skill:  Superb,
			Scale:  2,
			Traits: IntrinsicFast | IntrinsicFragile},
	}
	a["spider"] = entity.Assemblage{
		Metadata:      MetaTemplate(400, 6),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"spider", "chars:10"},
		CreatureComponent: &CreatureTemplate{
			Power:  Fair,
			Skill:  Great,
			Scale:  1,
			Traits: IntrinsicPoison},
	}
	a["killbot"] = entity.Assemblage{
		Metadata:      MetaTemplate(300, 8),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"killbot", "chars:11"},
		CreatureComponent: &CreatureTemplate{
			Power:  Superb,
			Skill:  Good,
			Scale:  0,
			Traits: IntrinsicElectrocute | IntrinsicDense | IntrinsicFragile},
	}
	a["ogre"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 8),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"ogre", "chars:15"},
		CreatureComponent: &CreatureTemplate{
			Power:  Great,
			Skill:  Fair,
			Scale:  3,
			Traits: NoIntrinsic},
	}
	a["boss1"] = entity.Assemblage{
		Metadata:      MetaTemplate(-1, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"elder spawn", "chars:5"},
		CreatureComponent: &CreatureTemplate{
			Power:  Legendary,
			Skill:  Superb,
			Scale:  5,
			Traits: IntrinsicEndboss},
	}

	a["globe"] = entity.Assemblage{
		Metadata:      MetaTemplate(30, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"health globe", "items:1"},
	}
	a["plantpot"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"plant pot", "items:3"},
		ItemComponent: &ItemTemplate{NoEquipSlot, 0, 0, 0, NoUse, NoItemTrait},
	}
	a["pistol"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"pistol", "items:4"},
		ItemComponent: &ItemTemplate{
			EquipmentSlot: GunEquipSlot,
			Durability:    12,
			WoundBonus:    1,
			DefenseBonus:  0,
			Use:           NoUse,
			Traits:        NoItemTrait},
	}
	a["machete"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"machete", "items:5"},
		ItemComponent: &ItemTemplate{
			EquipmentSlot: MeleeEquipSlot,
			Durability:    20,
			WoundBonus:    2,
			DefenseBonus:  0,
			Use:           NoUse,
			Traits:        NoItemTrait},
	}
	a["kevlar"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"kevlar armor", "items:6"},
		ItemComponent: &ItemTemplate{
			EquipmentSlot: ArmorEquipSlot,
			Durability:    20,
			WoundBonus:    0,
			DefenseBonus:  1,
			Use:           NoUse,
			Traits:        NoItemTrait},
	}
	a["medkit"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"medkit", "items:7"},
		ItemComponent: &ItemTemplate{NoEquipSlot, 0, 0, 0, MedkitUse, NoItemTrait},
	}
	a["monoblade"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 5),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"monoblade", "items:8"},
		ItemComponent: &ItemTemplate{
			EquipmentSlot: MeleeEquipSlot,
			Durability:    20,
			WoundBonus:    5,
			DefenseBonus:  0,
			Use:           NoUse,
			Traits:        NoItemTrait},
	}
	a["sledge"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 2),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"sledgehammer", "items:9"},
		ItemComponent: &ItemTemplate{
			EquipmentSlot: MeleeEquipSlot,
			Durability:    40,
			WoundBonus:    3,
			DefenseBonus:  0,
			Use:           NoUse,
			Traits:        ItemKnockBack},
	}
	a["spingun"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 4),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate{"spinner gun", "items:10"},
		ItemComponent: &ItemTemplate{
			EquipmentSlot: GunEquipSlot,
			Durability:    30,
			WoundBonus:    2,
			DefenseBonus:  0,
			Use:           NoUse,
			Traits:        ItemRapidFire},
	}
}
