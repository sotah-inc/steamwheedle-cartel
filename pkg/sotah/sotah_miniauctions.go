package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/petquality"
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
		if auc.Buyout > 0 {
			return float64(auc.Buyout) / float64(auc.Quantity)
		}

		if auc.UnitPrice > 0 {
			return float64(auc.UnitPrice)
		}

		return 0
	}()

	return miniAuction{
		auc.Item.Id,
		auc.Item.PetSpeciesId,
		auc.Item.PetQualityId,
		auc.Item.PetLevel,
		buyout,
		buyoutPer,
		auc.Quantity,
		auc.TimeLeft,
		[]blizzardv2.AuctionId{},
	}
}

type miniAuction struct {
	ItemId       blizzardv2.ItemId      `json:"itemId"`
	PetSpeciesId blizzardv2.PetId       `json:"pet_species_id"`
	PetQualityId petquality.PetQuality  `json:"pet_quality_id"`
	PetLevel     int                    `json:"pet_level"`
	Buyout       int64                  `json:"buyout"`
	BuyoutPer    float64                `json:"buyoutPer"`
	Quantity     int                    `json:"quantity"`
	TimeLeft     string                 `json:"timeLeft"`
	AucList      []blizzardv2.AuctionId `json:"aucList"`
}
