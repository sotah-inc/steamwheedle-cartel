package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
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
	foundName, err := item.BlizzardMeta.Name.Find(locale)
	if err != nil {
		return ShortItem{}, err
	}

	return ShortItem{
		SotahMeta: item.SotahMeta,
		Id:        item.BlizzardMeta.Id,
		Name:      foundName,
	}, nil
}

type ShortItem struct {
	SotahMeta ItemMeta `json:"sotah_meta"`

	Id   blizzardv2.ItemId `json:"id"`
	Name string            `json:"name"`
}
