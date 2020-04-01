package blizzardv2

type ItemIds []ItemId

type ItemIdsMap map[ItemId]struct{}

func (idsMap ItemIdsMap) ToItemIds() ItemIds {
	out := make(ItemIds, len(idsMap))
	i := 0
	for id := range idsMap {
		out[i] = id

		i += 1
	}

	return out
}

func (ids ItemIds) ToUniqueMap() ItemIdsMap {
	out := ItemIdsMap{}
	for _, id := range ids {
		out[id] = struct{}{}
	}

	return out
}

func (ids ItemIds) Merge(providedIds ItemIds) ItemIds {
	results := ids.ToUniqueMap()
	for _, id := range providedIds {
		results[id] = struct{}{}
	}

	return results.ToItemIds()
}
