package blizzardv2

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/binding"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/stattype"
)

type ValueDisplayStringTuple struct {
	Value         int            `json:"value"`
	DisplayString locale.Mapping `json:"display_string"`
}

type ItemSpell struct {
	Spell struct {
		Key  HrefReference  `json:"key"`
		Name locale.Mapping `json:"name"`
		Id   ItemSpellId    `json:"id"`
	} `json:"spell"`
	Description locale.Mapping `json:"description"`
}

type ItemColor struct {
	Red   int     `json:"r"`
	Green int     `json:"g"`
	Blue  int     `json:"b"`
	Alpha float32 `json:"a"`
}

type ItemStat struct {
	Type struct {
		Type stattype.StatType `json:"type"`
		Name locale.Mapping    `json:"name"`
	} `json:"type"`
	Value        int         `json:"value"`
	IsNegated    bool        `json:"is_negated"`
	Display      ItemDisplay `json:"display"`
	IsEquipBonus bool        `json:"is_equip_bonus"`
}

type ItemDisplay struct {
	DisplayString locale.Mapping `json:"display_string"`
	Color         ItemColor      `json:"color"`
}

type ItemValueDisplayStringTuple struct {
	Value   int         `json:"value"`
	Display ItemDisplay `json:"display"`
}

type ItemRecipeReagent struct {
	Reagent struct {
		Key  HrefReference  `json:"key"`
		Name locale.Mapping `json:"name"`
		Id   ItemId         `json:"id"`
	} `json:"reagent"`
	Quantity int `json:"quantity"`
}

type ItemSocket struct {
	SocketType struct {
		Type string         `json:"type"`
		Name locale.Mapping `json:"name"`
	} `json:"socket_type"`
}

type ItemRecipe struct {
	Reagents              []ItemRecipeReagent `json:"reagents"`
	ReagentsDisplayString locale.Mapping      `json:"reagents_display_string"`
}

type ItemRecipeWithItem struct {
	ItemRecipe

	Item ItemPreviewItemWithoutRecipeItem `json:"item"`
}

type ItemSetEffect struct {
	DisplayString locale.Mapping `json:"display_string"`
	RequiredCount int            `json:"required_count"`
}

type ItemSetItem struct {
	Item struct {
		Key  HrefReference  `json:"key"`
		Name locale.Mapping `json:"name"`
		Id   ItemId         `json:"id"`
	} `json:"item"`
}

type ItemPreviewItemBase struct {
	Item struct {
		Key HrefReference `json:"key"`
		Id  ItemId        `json:"id"`
	} `json:"item"`
	Quality       ItemQuality       `json:"quality"`
	Name          locale.Mapping    `json:"name"`
	Media         ItemMedia         `json:"media"`
	ItemClass     ItemClass         `json:"item_class"`
	ItemSubClass  ItemSubClass      `json:"item_subclass"`
	InventoryType ItemInventoryType `json:"inventory_type"`
	SellPrice     struct {
		Value          PriceValue `json:"value"`
		DisplayStrings struct {
			Header locale.Mapping `json:"header"`
			Gold   locale.Mapping `json:"gold"`
			Silver locale.Mapping `json:"silver"`
			Copper locale.Mapping `json:"copper"`
		} `json:"display_strings"`
	} `json:"sell_price"`

	// item-class-id: -1 (unknown)
	ShieldBlock      ItemValueDisplayStringTuple `json:"shield_block"`
	NameDescription  ItemDisplay                 `json:"name_description"`
	IsSubClassHidden bool                        `json:"is_subclass_hidden"`
	Description      locale.Mapping              `json:"description"`
	UniqueEquipped   locale.Mapping              `json:"unique_equipped"`

	// item-class-id: 0 (Consumable)
	Spells []ItemSpell `json:"spells"`

	// item-class-id: 1 (Container)
	ContainerSlots ValueDisplayStringTuple `json:"container_slots"`

	// item-class-id: 2 (Weapon)
	Weapon struct {
		Damage struct {
			MinValue      int            `json:"min_value"`
			MaxValue      int            `json:"max_value"`
			DisplayString locale.Mapping `json:"display_string"`
			DamageClass   struct {
				Type string         `json:"type"`
				Name locale.Mapping `json:"name"`
			} `json:"damage_class"`
		} `json:"damage"`
		AttackSpeed struct {
			Value         int            `json:"value"`
			DisplayString locale.Mapping `json:"display_string"`
		} `json:"attack_speed"`
		Dps struct {
			Value         float32        `json:"value"`
			DisplayString locale.Mapping `json:"display_string"`
		} `json:"dps"`
	} `json:"weapon"`

	// item-class-id: 3 (Gem)
	GemProperties struct {
		Effect       locale.Mapping `json:"effect"`
		MinItemLevel struct {
			Value         int            `json:"value"`
			DisplayString locale.Mapping `json:"display_string"`
		} `json:"min_item_level"`
	} `json:"gem_properties"`
	LimitCategory locale.Mapping `json:"limit_category"`

	// item-class-id: 4 (Armor)
	Binding struct {
		Type binding.Binding `json:"type"`
		Name locale.Mapping  `json:"name"`
	} `json:"binding"`
	Armor       ItemValueDisplayStringTuple `json:"armor"`
	Stats       []ItemStat                  `json:"stats"`
	Level       ItemValueDisplayStringTuple `json:"level"`
	Durability  ValueDisplayStringTuple     `json:"durability"`
	Sockets     []ItemSocket                `json:"sockets"`
	SocketBonus locale.Mapping              `json:"socket_bonus"`

	// item-class-id: 4 (Armor)
	// item-class-id: 9 (Recipe)
	Requirements struct {
		// item-class-id: 4 (Armor)
		Level ValueDisplayStringTuple `json:"level"`

		// item-class-id: 4 (Armor)
		// item-class-id: 9 (Recipe)
		Skill struct {
			Profession struct {
				Key  HrefReference  `json:"key"`
				Name locale.Mapping `json:"name"`
				Id   ProfessionId   `json:"id"`
			} `json:"profession"`
			Level         int            `json:"level"`
			DisplayString locale.Mapping `json:"display_string"`
		} `json:"skill"`

		// item-class-id: 4 (Armor)
		PlayableClasses struct {
			DisplayString locale.Mapping `json:"display_string"`
		} `json:"playable_classes"`

		// item-class-id: 15 (Misc)
		// item-sub-class-id: 5 (Mount)
		Ability struct {
			Spell struct {
				Key  HrefReference  `json:"key"`
				Name locale.Mapping `json:"name"`
				Id   ItemSpellId    `json:"id"`
			} `json:"spell"`
			DisplayString locale.Mapping `json:"display_string"`
		} `json:"ability"`
	} `json:"requirements"`

	// item-class-id: 4 (Armor)
	Set struct {
		ItemSet struct {
			Key  HrefReference  `json:"key"`
			Name locale.Mapping `json:"name"`
			Id   ItemSetId      `json:"id"`
		} `json:"item_set"`
		Items         []ItemSetItem   `json:"items"`
		Effects       []ItemSetEffect `json:"effects"`
		Legacy        locale.Mapping  `json:"legacy"`
		DisplayString locale.Mapping  `json:"display_string"`
	} `json:"set"`

	// item-class-id: 7 (Tradeskill)
	// item-subclass-id: 9 (Herb)
	CraftingReagent locale.Mapping `json:"crafting_reagent"`
}

type ItemPreviewItem struct {
	ItemPreviewItemBase

	// item-class-id: 9 (Recipe)
	Recipe ItemRecipeWithItem `json:"recipe"`
}

type ItemPreviewItemWithoutRecipeItem struct {
	ItemPreviewItemBase

	// item-class-id: 9 (Recipe)
	Recipe ItemRecipe `json:"recipe"`
}
