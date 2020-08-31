package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

// miniauction
func newMiniAuction(auc blizzardv2.Auction) miniAuction {
	buyout := func() int64 {
		if auc.Buyout > 0 {
			return auc.Buyout
		}

		if auc.UnitPrice > 0 {
			return auc.UnitPrice
		}

		return 0
	}()
	buyoutPer := func() float64 {
		if buyout == int64(0) {
			return 0
		}

		return float64(buyout) / float64(auc.Quantity)
	}()

	return miniAuction{
		auc.Item.Id,
		buyout,
		buyoutPer,
		auc.Quantity,
		auc.TimeLeft,
		[]blizzardv2.AuctionId{},
	}
}

type miniAuction struct {
	ItemId    blizzardv2.ItemId      `json:"itemId"`
	Buyout    int64                  `json:"buyout"`
	BuyoutPer float64                `json:"buyoutPer"`
	Quantity  int                    `json:"quantity"`
	TimeLeft  string                 `json:"timeLeft"`
	AucList   []blizzardv2.AuctionId `json:"aucList"`
}
