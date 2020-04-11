package blizzardv2

func NewItemBuyoutPerListMap(ids ItemIds) ItemBuyoutPerListMap {
	out := ItemBuyoutPerListMap{}
	for _, id := range ids {
		out[id] = ItemBuyoutPerList{}
	}

	return out
}

type ItemBuyoutPerListMap map[ItemId]ItemBuyoutPerList

func (perListMap ItemBuyoutPerListMap) Insert(id ItemId, buyoutPer float64) ItemBuyoutPerListMap {
	perListMap[id] = append(perListMap[id], buyoutPer)

	return perListMap
}
