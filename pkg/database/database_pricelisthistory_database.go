package database

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"github.com/boltdb/bolt"
)

type PricelistHistoryDatabase struct {
	db         *bolt.DB
	targetDate time.Time
}

func (phdBase PricelistHistoryDatabase) persistEncodedData(data map[blizzardv2.ItemId][]byte) error {
	logging.WithField("items", len(data)).Info("persisting encoded item-prices")

	err := phdBase.db.Batch(func(tx *bolt.Tx) error {
		for itemId, payload := range data {
			bkt, err := tx.CreateBucketIfNotExists(pricelistHistoryBucketName(itemId))
			if err != nil {
				return err
			}

			if err := bkt.Put(pricelistHistoryKeyName(), payload); err != nil {
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
