package items

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type PersistEncodedItemsInJob struct {
	Id                    blizzardv2.ItemId
	EncodedItem           []byte
	EncodedNormalizedName []byte
}

func (idBase Database) PersistEncodedItems(
	version gameversion.GameVersion,
	in chan PersistEncodedItemsInJob,
) (int, error) {
	logging.Info("persisting encoded items")

	totalPersisted := 0

	err := idBase.db.Batch(func(tx *bolt.Tx) error {
		itemsBucket, err := tx.CreateBucketIfNotExists(baseBucketName())
		if err != nil {
			return err
		}

		itemNamesBucket, err := tx.CreateBucketIfNotExists(namesBucketName())
		if err != nil {
			return err
		}

		for job := range in {
			tuple := blizzardv2.VersionItemTuple{
				GameVersion: version,
				Id:          job.Id,
			}
			if err := itemsBucket.Put(baseKeyName(tuple), job.EncodedItem); err != nil {
				return err
			}

			if err := itemNamesBucket.Put(nameKeyName(tuple), job.EncodedNormalizedName); err != nil {
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
