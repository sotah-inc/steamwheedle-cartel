package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
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
	foundName := item.BlizzardMeta.Name.FindOr(locale, "")
	foundQualityName := item.BlizzardMeta.Quality.Name.FindOr(locale, "")
	foundBinding := item.BlizzardMeta.PreviewItem.Binding.Name.FindOr(locale, "")
	foundHeader := item.BlizzardMeta.PreviewItem.SellPrice.DisplayStrings.Header.FindOr(locale, "")

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

type ShortItem struct {
	SotahMeta ItemMeta `json:"sotah_meta"`

	Id          blizzardv2.ItemId      `json:"id"`
	Name        string                 `json:"name"`
	Quality     ShortItemQuality       `json:"quality"`
	MaxCount    int                    `json:"max_count"`
	Level       int                    `json:"level"`
	ItemClassId blizzardv2.ItemClassId `json:"item_class_id"`
	Binding     string                 `json:"binding"`
	SellPrice   ShortItemSellPrice     `json:"sell_price"`
}
