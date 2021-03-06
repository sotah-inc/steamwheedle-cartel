package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewPricesFromEncoded(gzipEncoded []byte) (Prices, error) {
	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return Prices{}, err
	}

	var out Prices
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return Prices{}, err
	}

	return out, nil
}

type Prices struct {
	MinBuyoutPer         float64 `json:"min_buyout_per"`
	MaxBuyoutPer         float64 `json:"max_buyout_per"`
	AverageBuyoutPer     float64 `json:"average_buyout_per"`
	MedianBuyoutPer      float64 `json:"median_buyout_per"`
	MarketPriceBuyoutPer float64 `json:"market_price_buyout_per"`
	Volume               int64   `json:"volume"`
}

func (p Prices) EncodeForStorage() ([]byte, error) {
	jsonEncoded, err := json.Marshal(p)
	if err != nil {
		return []byte{}, err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return []byte{}, err
	}

	return gzipEncoded, nil
}

func (p Prices) ReceiveMiniAuction(mAuction miniAuction) Prices {
	if mAuction.BuyoutPer > 0 {
		if p.MinBuyoutPer == 0 || mAuction.BuyoutPer < p.MinBuyoutPer {
			p.MinBuyoutPer = mAuction.BuyoutPer
		}
		if p.MaxBuyoutPer == 0 || mAuction.BuyoutPer > p.MaxBuyoutPer {
			p.MaxBuyoutPer = mAuction.BuyoutPer
		}
	}

	p.Volume += int64(mAuction.Quantity * len(mAuction.AucList))

	return p
}

func (p Prices) ReceiveBuyoutPerSummary(buyoutPerSummary blizzardv2.ItemBuyoutPerSummary) Prices {
	p.AverageBuyoutPer = buyoutPerSummary.Average
	p.MedianBuyoutPer = buyoutPerSummary.Median
	p.MarketPriceBuyoutPer = buyoutPerSummary.MarketPrice

	return p
}
