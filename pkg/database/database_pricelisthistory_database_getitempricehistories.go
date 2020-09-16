package database

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (phdBase PricelistHistoryDatabase) getItemPrices(id blizzardv2.ItemId) (sotah.Prices, error) {
	out := sotah.Prices{}

	err := phdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(pricelistHistoryBucketName(id))
		if bkt == nil {
			return nil
		}

		value := bkt.Get(pricelistHistoryKeyName())
		if value == nil {
			return nil
		}

		var err error
		out, err = sotah.NewPricesFromEncoded(value)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.Prices{}, err
	}

	return out, nil

}
