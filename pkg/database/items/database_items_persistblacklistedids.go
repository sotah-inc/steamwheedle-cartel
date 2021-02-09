package items

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (idBase Database) PersistBlacklistedIds(
	ids []blizzardv2.ItemId,
) error {

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
