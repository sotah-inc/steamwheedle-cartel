package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/inventorytype"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemquality"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
)

func NewShortItemList(itemsList []Item, locale locale.Locale) (ShortItemList, error) {
	out := make(ShortItemList, len(itemsList))
	for i, item := range itemsList {
		resolvedShortItem, err := NewShortItem(item, locale)
		if err != nil {
			return ShortItemList{}, err
		}

		out[i] = resolvedShortItem
	}

	return out, nil
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

func NewShortItem(item Item, locale locale.Locale) (ShortItem, error) {
	foundName := item.BlizzardMeta.PreviewItem.Name.FindOr(locale, "")
	foundQualityName := item.BlizzardMeta.PreviewItem.Quality.Name.FindOr(locale, "")
	foundBinding := item.BlizzardMeta.PreviewItem.Binding.Name.FindOr(locale, "")
	foundHeader := item.BlizzardMeta.PreviewItem.SellPrice.DisplayStrings.Header.FindOr(locale, "")
	foundContainerSlots := item.BlizzardMeta.PreviewItem.ContainerSlots.DisplayString.FindOr(locale, "")
	foundDescription := item.BlizzardMeta.PreviewItem.Description.FindOr(locale, "")
	foundLevelRequirement := item.BlizzardMeta.PreviewItem.Requirements.Level.DisplayString.FindOr(locale, "")
	foundInventoryType := item.BlizzardMeta.PreviewItem.InventoryType.Name.FindOr(locale, "")
	foundItemSubclass := item.BlizzardMeta.ItemSubClass.Name.FindOr(locale, "")
	foundDurability := item.BlizzardMeta.PreviewItem.Durability.DisplayString.FindOr(locale, "")
	foundStats := make([]ShortItemStat, len(item.BlizzardMeta.PreviewItem.Stats))
	for i, stat := range item.BlizzardMeta.PreviewItem.Stats {
		foundStats[i] = ShortItemStat{
			DisplayValue: stat.Display.DisplayString.FindOr(locale, ""),
			IsNegated:    stat.IsNegated,
			Type:         stat.Type.Name.FindOr(locale, ""),
			Value:        stat.Value,
			IsEquipBonus: stat.IsEquipBonus,
		}
	}
	foundArmor := item.BlizzardMeta.PreviewItem.Armor.Display.DisplayString.FindOr(locale, "")
	foundSpells := make([]string, len(item.BlizzardMeta.PreviewItem.Spells))
	for i, spell := range item.BlizzardMeta.PreviewItem.Spells {
		foundSpells[i] = spell.Description.FindOr(locale, "")
	}
	foundSkillRequirement := item.BlizzardMeta.PreviewItem.Requirements.Skill.DisplayString.FindOr(locale, "")
	foundCraftingReagent := item.BlizzardMeta.PreviewItem.CraftingReagent.FindOr(locale, "")
	foundDamage := item.BlizzardMeta.PreviewItem.Weapon.Damage.DisplayString.FindOr(locale, "")
	foundAttackSpeed := item.BlizzardMeta.PreviewItem.Weapon.AttackSpeed.DisplayString.FindOr(locale, "")
	foundDps := item.BlizzardMeta.PreviewItem.Weapon.Dps.DisplayString.FindOr(locale, "")
	foundPlayableClasses := item.BlizzardMeta.PreviewItem.Requirements.PlayableClasses.DisplayString.FindOr(locale, "")
	foundSockets := make([]ShortItemSocket, len(item.BlizzardMeta.PreviewItem.Sockets))
	for i, socket := range item.BlizzardMeta.PreviewItem.Sockets {
		foundSockets[i] = ShortItemSocket{
			Type: socket.Type,
			Name: socket.Name.FindOr(locale, ""),
		}
	}
	foundSocketBonus := item.BlizzardMeta.PreviewItem.SocketBonus.FindOr(locale, "")
	foundUniqueEquipped := item.BlizzardMeta.PreviewItem.UniqueEquipped.FindOr(locale, "")
	foundReagentsDisplayString := item.BlizzardMeta.PreviewItem.Recipe.ReagentsDisplayString.FindOr(locale, "")

	return ShortItem{
		SotahMeta: item.SotahMeta,
		Id:        item.BlizzardMeta.Id,
		Name:      foundName,
		Quality: ShortItemQuality{
			Type: item.BlizzardMeta.Quality.Type,
			Name: foundQualityName,
		},
		MaxCount:    item.BlizzardMeta.MaxCount,
		Level:       item.BlizzardMeta.Level,
		ItemClassId: item.BlizzardMeta.ItemClass.Id,
		Binding:     foundBinding,
		SellPrice: ShortItemSellPrice{
			Value:  item.BlizzardMeta.PreviewItem.SellPrice.Value,
			Header: foundHeader,
		},
		ContainerSlots:   foundContainerSlots,
		Description:      foundDescription,
		LevelRequirement: foundLevelRequirement,
		InventoryType: ShortItemInventoryType{
			Type:          item.BlizzardMeta.InventoryType.Type,
			DisplayString: foundInventoryType,
		},
		ItemSubclass:          foundItemSubclass,
		Durability:            foundDurability,
		Stats:                 foundStats,
		Armor:                 foundArmor,
		Spells:                foundSpells,
		SkillRequirement:      foundSkillRequirement,
		ItemSubClassId:        item.BlizzardMeta.ItemSubClass.Id,
		CraftingReagent:       foundCraftingReagent,
		Damage:                foundDamage,
		AttackSpeed:           foundAttackSpeed,
		Dps:                   foundDps,
		PlayableClasses:       foundPlayableClasses,
		Sockets:               foundSockets,
		SocketBonus:           foundSocketBonus,
		UniqueEquipped:        foundUniqueEquipped,
		ReagentsDisplayString: foundReagentsDisplayString,
	}, nil
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

type ShortItem struct {
	SotahMeta ItemMeta `json:"sotah_meta"`

	Id                    blizzardv2.ItemId         `json:"id"`
	Name                  string                    `json:"name"`
	Quality               ShortItemQuality          `json:"quality"`
	MaxCount              int                       `json:"max_count"`
	Level                 int                       `json:"level"`
	ItemClassId           itemclass.Id              `json:"item_class_id"`
	Binding               string                    `json:"binding"`
	SellPrice             ShortItemSellPrice        `json:"sell_price"`
	ContainerSlots        string                    `json:"container_slots"`
	Description           string                    `json:"description"`
	LevelRequirement      string                    `json:"level_requirement"`
	InventoryType         ShortItemInventoryType    `json:"inventory_type"`
	ItemSubclass          string                    `json:"item_subclass"`
	Durability            string                    `json:"durability"`
	Stats                 []ShortItemStat           `json:"stats"`
	Armor                 string                    `json:"armor"`
	Spells                []string                  `json:"spells"`
	SkillRequirement      string                    `json:"skill_requirement"`
	ItemSubClassId        blizzardv2.ItemSubClassId `json:"item_subclass_id"`
	CraftingReagent       string                    `json:"crafting_reagent"`
	Damage                string                    `json:"damage"`
	AttackSpeed           string                    `json:"attack_speed"`
	Dps                   string                    `json:"dps"`
	PlayableClasses       string                    `json:"playable_classes"`
	Sockets               []ShortItemSocket         `json:"sockets"`
	SocketBonus           string                    `json:"socket_bonus"`
	UniqueEquipped        string                    `json:"unique_equipped"`
	ReagentsDisplayString string                    `json:"reagents_display_string"`
}
