package pricelisthistory

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func (phdBase Database) persistEncodedItemPrices(data map[blizzardv2.ItemId][]byte) error {
	logging.WithField("items", len(data)).Info("persisting encoded item-prices")

	err := phdBase.db.Batch(func(tx *bolt.Tx) error {
		for itemId, payload := range data {
			bkt, err := tx.CreateBucketIfNotExists(baseBucketName(itemId))
			if err != nil {
				return err
			}

			if err := bkt.Put(baseKeyName(), payload); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
