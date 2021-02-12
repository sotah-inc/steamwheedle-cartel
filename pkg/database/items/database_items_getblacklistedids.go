package items

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (idBase Database) GetBlacklistedIds() (blizzardv2.ItemIds, error) {
	var out blizzardv2.ItemIds

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blacklistBucketName())
		if bkt == nil {
			return nil
		}

		return bkt.ForEach(func(k []byte, v []byte) error {
			id, err := itemIdFromBlacklistKeyName(k)
			if err != nil {
				return err
			}

			out = append(out, id)

			return nil
		})
	})
	if err != nil {
		return blizzardv2.ItemIds{}, err
	}

	return out, nil
}
