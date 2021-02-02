package blizzardv2

func NewItemBuyoutPerSummaryMap(perListmap ItemBuyoutPerListMap) ItemBuyoutPerSummaryMap {
	out := ItemBuyoutPerSummaryMap{}
	for itemId, perList := range perListmap {
		out[itemId] = NewItemBuyoutPerSummary(perList)
	}

	return out
}

type ItemBuyoutPerSummaryMap map[ItemId]ItemBuyoutPerSummary

func NewItemBuyoutPerSummary(perList ItemBuyoutPerList) ItemBuyoutPerSummary {
	return ItemBuyoutPerSummary{
		Average:     perList.Average(),
		Median:      perList.Median(),
		MarketPrice: perList.MarketPrice(),
	}
}

type ItemBuyoutPerSummary struct {
	Average     float64
	Median      float64
	MarketPrice float64
}
