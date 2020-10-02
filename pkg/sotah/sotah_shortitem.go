package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
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
	foundDurability := item.BlizzardMeta.PreviewItem.Durability.Display.DisplayString.FindOr(locale, "")
	foundStats := make([]ShortItemStat, len(item.BlizzardMeta.PreviewItem.Stats))
	for i, stat := range item.BlizzardMeta.PreviewItem.Stats {
		foundStats[i] = ShortItemStat{
			DisplayValue: stat.Display.DisplayString.FindOr(locale, ""),
			IsNegated:    stat.IsNegated,
			Type:         stat.Type.Name.FindOr(locale, ""),
			Value:        stat.Value,
		}
	}
	foundArmor := item.BlizzardMeta.PreviewItem.Armor.Display.DisplayString.FindOr(locale, "")

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
		InventoryType:    foundInventoryType,
		ItemSubclass:     foundItemSubclass,
		Durability:       foundDurability,
		Stats:            foundStats,
		Armor:            foundArmor,
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
}

type ShortItem struct {
	SotahMeta ItemMeta `json:"sotah_meta"`

	Id               blizzardv2.ItemId  `json:"id"`
	Name             string             `json:"name"`
	Quality          ShortItemQuality   `json:"quality"`
	MaxCount         int                `json:"max_count"`
	Level            int                `json:"level"`
	ItemClassId      itemclass.Id       `json:"item_class_id"`
	Binding          string             `json:"binding"`
	SellPrice        ShortItemSellPrice `json:"sell_price"`
	ContainerSlots   string             `json:"container_slots"`
	Description      string             `json:"description"`
	LevelRequirement string             `json:"level_requirement"`
	InventoryType    string             `json:"inventory_type"`
	ItemSubclass     string             `json:"item_subclass"`
	Durability       string             `json:"durability"`
	Stats            []ShortItemStat    `json:"stats"`
	Armor            string             `json:"armor"`
}
