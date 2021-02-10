package items

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func (idBase Database) PersistBlacklistedIds(
	ids []blizzardv2.ItemId,
) error {
	logging.WithField("erroneous-ids", ids).Info("persisting blacklisted item-ids")

	return idBase.db.Batch(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(blacklistBucketName())
		if err != nil {
			return err
		}

		for _, id := range ids {
			if err := bkt.Put(blacklistKeyName(id), []byte{}); err != nil {
				return err
			}
		}

		return nil
	})
}
