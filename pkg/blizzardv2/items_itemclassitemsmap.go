package blizzardv2

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"
)

func NewItemClassItemsMap(ids []itemclass.Id) ItemClassItemsMap {
	out := ItemClassItemsMap{}
	for _, id := range ids {
		out[id] = ItemIds{}
	}

	return out
}

type ItemClassItemsMap map[itemclass.Id]ItemIds

func (iciMap ItemClassItemsMap) Find(classId itemclass.Id) ItemIds {
	found, ok := iciMap[classId]
	if !ok {
		return ItemIds{}
	}

	return found
}

func (iciMap ItemClassItemsMap) Insert(
	providedClassId itemclass.Id,
	providedItemId ItemId,
) ItemClassItemsMap {
	iciMap[providedClassId] = iciMap.Find(providedClassId).Merge(ItemIds{providedItemId})

	return iciMap
}

func (iciMap ItemClassItemsMap) ItemClassIds() []itemclass.Id {
	out := make([]itemclass.Id, len(iciMap))
	i := 0
	for id := range iciMap {
		out[i] = id

		i += 1
	}

	return out
}
