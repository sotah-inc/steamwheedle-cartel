package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func NewItemPricesFromMiniAuctionList(maList MiniAuctionList) ItemPrices {
	out := NewItemPrices(maList.ItemIds())
	for _, mAuction := range maList {
		out = out.ReceiveMiniAuction(mAuction)
	}

	for id, summary := range maList.ToItemBuyoutPerListMap() {
		out = out.ReceiveBuyoutPerSummary(id, summary)
	}

	return out
}

func NewItemPrices(ids blizzardv2.ItemIds) ItemPrices {
	out := ItemPrices{}
	for _, id := range ids {
		out[id] = Prices{}
	}

	return out
}

type ItemPrices map[blizzardv2.ItemId]Prices

func (iPrices ItemPrices) ReceiveMiniAuction(mAuction miniAuction) ItemPrices {
	iPrices[mAuction.ItemId] = iPrices[mAuction.ItemId].ReceiveMiniAuction(mAuction)

	return iPrices
}
func (iPrices ItemPrices) ReceiveBuyoutPerSummary(
	id blizzardv2.ItemId,
	buyoutPerSummary blizzardv2.ItemBuyoutPerSummary,
) ItemPrices {
	iPrices[id] = iPrices[id].ReceiveBuyoutPerSummary(buyoutPerSummary)

	return iPrices
}

func (iPrices ItemPrices) ItemIds() []blizzardv2.ItemId {
	out := make([]blizzardv2.ItemId, len(iPrices))
	i := 0
	for id := range iPrices {
		out[i] = id

		i += 1
	}

	return out
}

type ItemBuyoutPers map[blizzardv2.ItemId][]float64
