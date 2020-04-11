package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

// miniauction
func newMiniAuction(auc blizzardv2.Auction) miniAuction {
	var buyoutPer float32
	if auc.Buyout > 0 {
		buyoutPer = float32(auc.Buyout) / float32(auc.Quantity)
	}

	return miniAuction{
		auc.Item.Id,
		auc.Buyout,
		buyoutPer,
		auc.Quantity,
		auc.TimeLeft,
		[]blizzardv2.AuctionId{},
	}
}

type miniAuction struct {
	ItemId    blizzardv2.ItemId      `json:"itemId"`
	Buyout    int64                  `json:"buyout"`
	BuyoutPer float32                `json:"buyoutPer"`
	Quantity  int                    `json:"quantity"`
	TimeLeft  string                 `json:"timeLeft"`
	AucList   []blizzardv2.AuctionId `json:"aucList"`
}
