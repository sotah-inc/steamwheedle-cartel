package items

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
)

func (idBase Database) PruneItemBlacklist(version gameversion.GameVersion) error {
	return idBase.db.Batch(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blacklistBucketName(version))
		if bkt == nil {
			return nil
		}

		return bkt.ForEach(func(k []byte, v []byte) error {
			return bkt.Delete(k)
		})
	})
}
