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
		Metadata:           MetaTemplate(-1, 0),
		PosComponent:       PosTemplate(),
		NameComponent:      NameTemplate("protagonist", "chars:16", PronounIt, false),
		MutationsComponent: entity.NewDefaultTemplate((*Mutations)(nil), MutationsComponent, nil),
		FixedInventoryComponent: entity.NewDefaultTemplate((*FixedInventory)(nil), FixedInventoryComponent, map[string]interface{}{
			"Ammo":     50,
			"MedKits":  0,
			"Grenades": 0,
			"Armor":    0}),
		CreatureComponent: &CreatureTemplate{
			Attack1:    WeaponFist,
			Attack2:    WeaponPistol,
			Hp:         100,
			Intrinsics: IntrinsicMartialArtist},
	}
	a["infectedHuman"] = entity.Assemblage{
		Metadata:      MetaTemplate(100, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("infected human", "chars:13", PronounHe, false),
		CreatureComponent: &CreatureTemplate{
			Attack1:    WeaponFist,
			Hp:         20,
			Intrinsics: IntrinsicMartialArtist},
	}
	a["infectedGuard"] = entity.Assemblage{
		Metadata:      MetaTemplate(100, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("infected guard", "chars:12", PronounHe, false),
		CreatureComponent: &CreatureTemplate{
			Attack1:    WeaponFist,
			Attack2:    WeaponPistol,
			Hp:         30,
			Intrinsics: IntrinsicMartialArtist},
	}
	a["zombie"] = entity.Assemblage{
		Metadata:      MetaTemplate(100, 3),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("zombie", "chars:1", PronounIt, false),
		CreatureComponent: &CreatureTemplate{
			Attack1:    WeaponFist,
			Hp:         25,
			Intrinsics: NoIntrinsic},
	}
	a["dogthing"] = entity.Assemblage{
		Metadata:      MetaTemplate(150, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("dog-thing", "chars:2", PronounIt, false),
		CreatureComponent: &CreatureTemplate{
			Attack1:    WeaponJaws,
			Hp:         15,
			Intrinsics: NoIntrinsic},
	}
	a["belcher"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 4),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("belcher", "chars:3", PronounIt, false),
		CreatureComponent: &CreatureTemplate{
			Attack1:    WeaponClaw,
			Attack2:    WeaponBile,
			Hp:         30,
			Intrinsics: NoIntrinsic},
	}
	a["crawlingmass"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 4),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("crawling mass", "chars:7", PronounIt, false),
		CreatureComponent: &CreatureTemplate{
			Attack1:    WeaponCrawl,
			Hp:         70,
			Intrinsics: IntrinsicSlow | IntrinsicDeathsplode},
	}
	a["cyclops"] = entity.Assemblage{
		Metadata:      MetaTemplate(300, 4),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("cyclops", "chars:6", PronounIt, false),
		CreatureComponent: &CreatureTemplate{
			Attack1:    WeaponGaze,
			Attack2:    WeaponPsiBlast,
			Hp:         30,
			Intrinsics: NoIntrinsic},
	}
	a["wendigo"] = entity.Assemblage{
		Metadata:      MetaTemplate(300, 6),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("wendigo", "chars:8", PronounIt, false),
		CreatureComponent: &CreatureTemplate{
			Attack1:    WeaponClaw,
			Hp:         50,
			Intrinsics: IntrinsicFast},
	}
	a["spider"] = entity.Assemblage{
		Metadata:      MetaTemplate(400, 6),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("spider", "chars:10", PronounIt, false),
		CreatureComponent: &CreatureTemplate{
			Attack1: WeaponSpider,
			// TODO: Web spraying weapon?
			Hp:         50,
			Intrinsics: NoIntrinsic},
	}
	a["infectedSolder"] = entity.Assemblage{
		Metadata:      MetaTemplate(100, 8),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("infected soldier", "chars:14", PronounHe, false),
		CreatureComponent: &CreatureTemplate{
			Attack1:    WeaponBayonet,
			Attack2:    WeaponRifle,
			Hp:         45,
			Intrinsics: IntrinsicMartialArtist},
	}
	a["killbot"] = entity.Assemblage{
		Metadata:      MetaTemplate(300, 8),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("killbot", "chars:11", PronounIt, false),
		CreatureComponent: &CreatureTemplate{
			Attack1:    WeaponSaw,
			Attack2:    WeaponZap,
			Hp:         100,
			Intrinsics: IntrinsicUnliving},
	}
	a["ogre"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 8),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("ogre", "chars:15", PronounIt, false),
		CreatureComponent: &CreatureTemplate{
			Attack1:    WeaponSmash,
			Hp:         120,
			Intrinsics: NoIntrinsic},
	}
	a["boss1"] = entity.Assemblage{
		Metadata:      MetaTemplate(-1, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("elder spawn", "chars:5", PronounIt, false),
		CreatureComponent: &CreatureTemplate{
			Attack1:    WeaponSmash,
			Attack2:    WeaponNether,
			Hp:         200,
			Intrinsics: IntrinsicEndboss},
	}

	a["blood_small"] = entity.Assemblage{
		Metadata:       MetaTemplate(-1, 0),
		PosComponent:   PosTemplate(),
		NameComponent:  NameTemplate("blood splatter", "items:16", PronounIt, false),
		DecalComponent: DecalTemplate(0),
	}
	a["blood_large"] = entity.Assemblage{
		Metadata:       MetaTemplate(-1, 0),
		PosComponent:   PosTemplate(),
		NameComponent:  NameTemplate("blood pool", "items:15", PronounIt, false),
		DecalComponent: DecalTemplate(0),
	}
	a["blood_trail"] = entity.Assemblage{
		Metadata:       MetaTemplate(-1, 0),
		PosComponent:   PosTemplate(),
		NameComponent:  NameTemplate("bloody trail", "items:17", PronounIt, false),
		DecalComponent: DecalTemplate(0),
	}

	a["globe"] = entity.Assemblage{
		Metadata:       MetaTemplate(30, 0),
		PosComponent:   PosTemplate(),
		NameComponent:  NameTemplate("globe", "items:1", PronounIt, false),
		DecalComponent: DecalTemplate(0),
	}

	a["machete"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("machete", "items:5", PronounIt, false),
		ItemComponent: &ItemTemplate{
			EquipmentSlot: MeleeEquipSlot,
			Durability:    20,
			WoundBonus:    2,
			DefenseBonus:  0,
			Use:           NoUse,
			Traits:        NoItemTrait},
	}
	a["vest"] = entity.Assemblage{
		Metadata:      MetaTemplate(400, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("tactical vest", "items:6", PronounIt, false),
		ItemComponent: &ItemTemplate{
			EquipmentSlot: ArmorEquipSlot,
			Durability:    20,
			WoundBonus:    0,
			DefenseBonus:  VestArmorLevel,
			Use:           NoUse,
			Traits:        NoItemTrait},
	}
	a["riot"] = entity.Assemblage{
		Metadata:      MetaTemplate(400, 3),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("riot armor", "items:13", PronounIt, false),
		ItemComponent: &ItemTemplate{
			EquipmentSlot: ArmorEquipSlot,
			Durability:    20,
			WoundBonus:    0,
			DefenseBonus:  RiotArmorLevel,
			Use:           NoUse,
			Traits:        NoItemTrait},
	}
	a["hardsuit"] = entity.Assemblage{
		Metadata:      MetaTemplate(600, 8),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("hard suit", "items:14", PronounIt, false),
		ItemComponent: &ItemTemplate{
			EquipmentSlot: ArmorEquipSlot,
			Durability:    40,
			WoundBonus:    0,
			DefenseBonus:  HardsuitArmorLevel,
			Use:           NoUse,
			Traits:        ItemHardsuit},
	}
	a["medkit"] = entity.Assemblage{
		Metadata:      MetaTemplate(300, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("medkit", "items:7", PronounIt, false),
		ItemComponent: &ItemTemplate{NoEquipSlot, 0, 0, 0, MedkitUse, NoItemTrait},
	}
	//	a["stabilizer"] = entity.Assemblage{
	//		Metadata:      MetaTemplate(300, 0),
	//		PosComponent:  PosTemplate(),
	//		NameComponent: NameTemplate("genetic stabilizer", "items:18", PronounIt, false),
	//		ItemComponent: &ItemTemplate{NoEquipSlot, 0, 0, 0, StabilizerUse, NoItemTrait},
	//	}
	a["monoblade"] = entity.Assemblage{
		Metadata:      MetaTemplate(200, 5),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("monoblade", "items:8", PronounIt, false),
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
		NameComponent: NameTemplate("sledgehammer", "items:9", PronounIt, false),
		ItemComponent: &ItemTemplate{
			EquipmentSlot: MeleeEquipSlot,
			Durability:    40,
			WoundBonus:    3,
			DefenseBonus:  0,
			Use:           NoUse,
			Traits:        ItemKnockback},
	}
	a["ammo"] = entity.Assemblage{
		Metadata:      MetaTemplate(150, 0),
		PosComponent:  PosTemplate(),
		NameComponent: NameTemplate("ammo clip", "items:19", PronounIt, false),
		ItemComponent: &ItemTemplate{NoEquipSlot, 0, 0, 0, AmmoUse, NoItemTrait},
	}
}
