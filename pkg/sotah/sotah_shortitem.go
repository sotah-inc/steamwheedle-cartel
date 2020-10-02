package sotah

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemquality"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
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
	foundName, err := item.BlizzardMeta.Name.Find(locale)
	if err != nil {
		return ShortItem{}, err
	}

	foundQualityName, err := item.BlizzardMeta.Quality.Name.Find(locale)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error":   err.Error(),
			"item_id": item.BlizzardMeta.Id,
			"locale":  locale,
		}).Error("failed to resolve item.BlizzardMeta.Quality.Name")

		return ShortItem{}, err
	}

	foundBinding, err := item.BlizzardMeta.PreviewItem.Binding.Name.Find(locale)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error":   err.Error(),
			"item_id": item.BlizzardMeta.Id,
			"locale":  locale,
		}).Error("failed to resolve item.BlizzardMeta.PreviewItem.Binding.Name")

		return ShortItem{}, err
	}

	foundHeader, err := item.BlizzardMeta.PreviewItem.SellPrice.DisplayStrings.Header.Find(locale)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error":   err.Error(),
			"item_id": item.BlizzardMeta.Id,
			"locale":  locale,
		}).Error("failed to resolve preview item.BlizzardMeta.PreviewItem.SellPrice.DisplayStrings.Header")

		return ShortItem{}, err
	}

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
