package database

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type PersistEncodedItemsInJob struct {
	Id              blizzardv2.ItemId
	GzipEncodedData []byte
}

func (idBase ItemsDatabase) PersistEncodedItems(
	in chan PersistEncodedItemsInJob,
	idNameMap sotah.ItemIdNameMap,
) error {
	logging.Info("Persisting encoded items")

	err := idBase.db.Batch(func(tx *bolt.Tx) error {
		itemsBucket, err := tx.CreateBucketIfNotExists(databaseItemsBucketName())
		if err != nil {
			return err
		}

		itemNamesBucket, err := tx.CreateBucketIfNotExists(databaseItemNamesBucketName())
		if err != nil {
			return err
		}

		for job := range in {
			if err := itemsBucket.Put(itemsKeyName(job.Id), job.GzipEncodedData); err != nil {
				return err
			}
		}

		i := 0
		for id, normalizedName := range idNameMap {
			if normalizedName.IsZero() {
				continue
			}

			encodedNormalizedName, err := normalizedName.EncodeForStorage()
			if err != nil {
				return err
			}

			if err := itemNamesBucket.Put(itemNameKeyName(id), encodedNormalizedName); err != nil {
				return err
			}

			if i%100 == 0 {
				logging.WithFields(logrus.Fields{
					"id":   id,
					"name": normalizedName,
					"i":    i,
				}).Info("Inserted into item-names bucket")
			}

			i++
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
