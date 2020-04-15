package database

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	"github.com/boltdb/bolt"
)

func (idBase ItemsDatabase) PersistItems(in chan sotah.Item) (int, error) {
	totalPersisted := 0

	err := idBase.db.Batch(func(tx *bolt.Tx) error {
		itemsBucket, err := tx.CreateBucketIfNotExists(databaseItemsBucketName())
		if err != nil {
			return err
		}

		itemNamesBucket, err := tx.CreateBucketIfNotExists(databaseItemNamesBucketName())
		if err != nil {
			return err
		}

		for item := range in {
			encodedItem, err := item.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"item-id": item.BlizzardMeta.Id,
					"error":   err.Error(),
				}).Error("failed to encode item")

				return err
			}

			if err := itemsBucket.Put(itemsKeyName(item.BlizzardMeta.Id), encodedItem); err != nil {
				return err
			}

			encodedNormalizedName, err := item.SotahMeta.NormalizedName.EncodeForStorage()
			if err != nil {
				return err
			}

			if err := itemNamesBucket.Put(itemNameKeyName(item.BlizzardMeta.Id), encodedNormalizedName); err != nil {
				return err
			}

			totalPersisted += 1
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return totalPersisted, nil
}
