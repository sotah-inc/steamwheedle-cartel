package items

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func (idBase Database) GetBlacklistedIds() (blizzardv2.ItemIds, error) {
	logging.Info("Database.GetBlacklistedIds()")

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

	logging.WithField("blacklisted-ids", out).Info("found blacklisted-ids")

	return out, nil
}
