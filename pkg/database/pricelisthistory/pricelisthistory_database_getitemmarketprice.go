package pricelisthistory

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (phdBase Database) getItemMarketPrice(id blizzardv2.ItemId) (float64, error) {
	out := float64(0)

	err := phdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName(id))
		if bkt == nil {
			return nil
		}

		value := bkt.Get(baseKeyName())
		if value == nil {
			return nil
		}

		prices, err := sotah.NewPricesFromEncoded(value)
		if err != nil {
			return err
		}

		out = prices.MarketPriceBuyoutPer

		return nil
	})
	if err != nil {
		return 0, err
	}

	return out, nil
}
