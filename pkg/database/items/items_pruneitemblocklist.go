package items

import (
	"github.com/boltdb/bolt"
)

func (idBase Database) PruneItemBlacklist() error {
	return idBase.db.Batch(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blacklistBucketName())
		if bkt == nil {
			return nil
		}

		return bkt.ForEach(func(k []byte, v []byte) error {
			return bkt.Delete(k)
		})
	})
}
