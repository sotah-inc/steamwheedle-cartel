package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/inventorytype"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemquality"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
)

func NewShortItemList(itemsList []Item, locale locale.Locale) ShortItemList {
	out := make(ShortItemList, len(itemsList))
	for i, item := range itemsList {
		out[i] = NewShortItem(item, locale)
	}

	return out
}

type ShortItemInventoryType struct {
	Type          inventorytype.InventoryType `json:"type"`
	DisplayString string                      `json:"display_string"`
}

type ShortItemSocket struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type ShortItemList []ShortItem

func NewShortItem(item Item, locale locale.Locale) ShortItem {
	foundReagentsDisplayString := item.BlizzardMeta.PreviewItem.Recipe.ReagentsDisplayString.FindOr(locale, "")
	foundLevel := item.BlizzardMeta.PreviewItem.Recipe.Item.Level.Value
	if foundLevel == 0 {
		foundLevel = item.BlizzardMeta.Level
	}

	return ShortItem{
		ShortItemBase: NewShortItemFromPreviewItem(ShortItemParams{
			previewItem: item.BlizzardMeta.PreviewItem.ItemPreviewItemBase,
			locale:      locale,
			sotahMeta:   item.SotahMeta,
			id:          item.BlizzardMeta.Id,
			maxCount:    item.BlizzardMeta.MaxCount,
			level:       item.BlizzardMeta.Level,
		}),
		RecipeItem: ShortItemWithoutRecipeItem{
			ShortItemBase: NewShortItemFromPreviewItem(ShortItemParams{
				previewItem: item.BlizzardMeta.PreviewItem.Recipe.Item.ItemPreviewItemBase,
				locale:      locale,
				sotahMeta:   item.SotahMeta,
				id:          item.BlizzardMeta.PreviewItem.Recipe.Item.ItemPreviewItemBase.Item.Id,
				maxCount:    0,
				level:       foundLevel,
			}),
		},
		ReagentsDisplayString: foundReagentsDisplayString,
	}
}

type ShortItemParams struct {
	previewItem blizzardv2.ItemPreviewItemBase
	locale      locale.Locale
	sotahMeta   ItemMeta
	id          blizzardv2.ItemId
	maxCount    int
	level       int
}

func NewShortItemFromPreviewItem(params ShortItemParams) ShortItemBase {
	foundName := params.previewItem.Name.FindOr(params.locale, "")
	foundQualityName := params.previewItem.Quality.Name.FindOr(params.locale, "")
	foundBinding := params.previewItem.Binding.Name.FindOr(params.locale, "")
	foundHeader := params.previewItem.SellPrice.DisplayStrings.Header.FindOr(params.locale, "")
	foundContainerSlots := params.previewItem.ContainerSlots.DisplayString.FindOr(params.locale, "")
	foundDescription := params.previewItem.Description.FindOr(params.locale, "")
	foundLevelRequirement := params.previewItem.Requirements.Level.DisplayString.FindOr(params.locale, "")
	foundInventoryType := params.previewItem.InventoryType.Name.FindOr(params.locale, "")
	foundItemSubclass := params.previewItem.ItemSubClass.Name.FindOr(params.locale, "")
	foundDurability := params.previewItem.Durability.DisplayString.FindOr(params.locale, "")
	foundStats := make([]ShortItemStat, len(params.previewItem.Stats))
	for i, stat := range params.previewItem.Stats {
		foundStats[i] = ShortItemStat{
			DisplayValue: stat.Display.DisplayString.FindOr(params.locale, ""),
			IsNegated:    stat.IsNegated,
			Type:         stat.Type.Name.FindOr(params.locale, ""),
			Value:        stat.Value,
			IsEquipBonus: stat.IsEquipBonus,
		}
	}
	foundArmor := params.previewItem.Armor.Display.DisplayString.FindOr(params.locale, "")
	foundSpells := make([]string, len(params.previewItem.Spells))
	for i, spell := range params.previewItem.Spells {
		foundSpells[i] = spell.Description.FindOr(params.locale, "")
	}
	foundSkillRequirement := params.previewItem.Requirements.Skill.DisplayString.FindOr(params.locale, "")
	foundCraftingReagent := params.previewItem.CraftingReagent.FindOr(params.locale, "")
	foundDamage := params.previewItem.Weapon.Damage.DisplayString.FindOr(params.locale, "")
	foundAttackSpeed := params.previewItem.Weapon.AttackSpeed.DisplayString.FindOr(params.locale, "")
	foundDps := params.previewItem.Weapon.Dps.DisplayString.FindOr(params.locale, "")
	foundPlayableClasses := params.previewItem.Requirements.PlayableClasses.DisplayString.FindOr(params.locale, "")
	foundSockets := make([]ShortItemSocket, len(params.previewItem.Sockets))
	for i, socket := range params.previewItem.Sockets {
		foundSockets[i] = ShortItemSocket{
			Type: socket.SocketType.Type,
			Name: socket.SocketType.Name.FindOr(params.locale, ""),
		}
	}
	foundSocketBonus := params.previewItem.SocketBonus.FindOr(params.locale, "")
	foundUniqueEquipped := params.previewItem.UniqueEquipped.FindOr(params.locale, "")
	foundGemEffect := params.previewItem.GemProperties.Effect.FindOr(params.locale, "")
	foundGemMinItemLevel := params.previewItem.GemProperties.MinItemLevel.DisplayString.FindOr(params.locale, "")
	foundAbilityRequirement := params.previewItem.Requirements.Ability.DisplayString.FindOr(params.locale, "")
	foundLimitCategory := params.previewItem.LimitCategory.FindOr(params.locale, "")
	foundNameDescription := params.previewItem.NameDescription.DisplayString.FindOr(params.locale, "")

	return ShortItemBase{
		SotahMeta: params.sotahMeta,
		Id:        params.id,
		Name:      foundName,
		Quality: ShortItemQuality{
			Type: params.previewItem.Quality.Type,
			Name: foundQualityName,
		},
		MaxCount:    params.maxCount,
		Level:       params.level,
		ItemClassId: params.previewItem.ItemClass.Id,
		Binding:     foundBinding,
		SellPrice: ShortItemSellPrice{
			Value:  params.previewItem.SellPrice.Value,
			Header: foundHeader,
		},
		ContainerSlots:   foundContainerSlots,
		Description:      foundDescription,
		LevelRequirement: foundLevelRequirement,
		InventoryType: ShortItemInventoryType{
			Type:          params.previewItem.InventoryType.Type,
			DisplayString: foundInventoryType,
		},
		ItemSubclass:       foundItemSubclass,
		Durability:         foundDurability,
		Stats:              foundStats,
		Armor:              foundArmor,
		Spells:             foundSpells,
		SkillRequirement:   foundSkillRequirement,
		ItemSubClassId:     params.previewItem.ItemSubClass.Id,
		CraftingReagent:    foundCraftingReagent,
		Damage:             foundDamage,
		AttackSpeed:        foundAttackSpeed,
		Dps:                foundDps,
		PlayableClasses:    foundPlayableClasses,
		Sockets:            foundSockets,
		SocketBonus:        foundSocketBonus,
		UniqueEquipped:     foundUniqueEquipped,
		GemEffect:          foundGemEffect,
		GemMinItemLevel:    foundGemMinItemLevel,
		AbilityRequirement: foundAbilityRequirement,
		LimitCategory:      foundLimitCategory,
		NameDescription:    foundNameDescription,
	}
}

type ShortItemQuality struct {
	Type itemquality.ItemQuality `json:"type"`
	Name string                  `json:"name"`
}

type ShortItemSellPrice struct {
	Value  blizzardv2.PriceValue `json:"value"`
	Header string                `json:"header"`
}

type ShortItemStat struct {
	DisplayValue string `json:"display_value"`
	IsNegated    bool   `json:"is_negated"`
	Type         string `json:"type"`
	Value        int    `json:"value"`
	IsEquipBonus bool   `json:"is_equip_bonus"`
}

type ShortItemRecipeItem struct {
	Item ShortItem `json:"item"`
}

type ShortItemBase struct {
	SotahMeta ItemMeta `json:"sotah_meta"`

	Id                 blizzardv2.ItemId         `json:"id"`
	Name               string                    `json:"name"`
	Quality            ShortItemQuality          `json:"quality"`
	MaxCount           int                       `json:"max_count"`
	Level              int                       `json:"level"`
	ItemClassId        itemclass.Id              `json:"item_class_id"`
	Binding            string                    `json:"binding"`
	SellPrice          ShortItemSellPrice        `json:"sell_price"`
	ContainerSlots     string                    `json:"container_slots"`
	Description        string                    `json:"description"`
	LevelRequirement   string                    `json:"level_requirement"`
	InventoryType      ShortItemInventoryType    `json:"inventory_type"`
	ItemSubclass       string                    `json:"item_subclass"`
	Durability         string                    `json:"durability"`
	Stats              []ShortItemStat           `json:"stats"`
	Armor              string                    `json:"armor"`
	Spells             []string                  `json:"spells"`
	SkillRequirement   string                    `json:"skill_requirement"`
	ItemSubClassId     blizzardv2.ItemSubClassId `json:"item_subclass_id"`
	CraftingReagent    string                    `json:"crafting_reagent"`
	Damage             string                    `json:"damage"`
	AttackSpeed        string                    `json:"attack_speed"`
	Dps                string                    `json:"dps"`
	PlayableClasses    string                    `json:"playable_classes"`
	Sockets            []ShortItemSocket         `json:"sockets"`
	SocketBonus        string                    `json:"socket_bonus"`
	UniqueEquipped     string                    `json:"unique_equipped"`
	GemEffect          string                    `json:"gem_effect"`
	GemMinItemLevel    string                    `json:"gem_min_item_level"`
	AbilityRequirement string                    `json:"ability_requirement"`
	LimitCategory      string                    `json:"limit_category"`
	NameDescription    string                    `json:"name_description"`
}

type ShortItem struct {
	ShortItemBase

	RecipeItem            ShortItemWithoutRecipeItem `json:"recipe_item"`
	ReagentsDisplayString string                     `json:"reagents_display_string"`
}

type ShortItemWithoutRecipeItem struct {
	ShortItemBase
}
