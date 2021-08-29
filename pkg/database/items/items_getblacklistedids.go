package items

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
)

func (idBase Database) GetBlacklistedIds(
	version gameversion.GameVersion,
) (blizzardv2.ItemIds, error) {
	var out blizzardv2.ItemIds

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blacklistBucketName())
		if bkt == nil {
			return nil
		}

		return bkt.ForEach(func(k []byte, v []byte) error {
			tuple, err := tupleFromBlacklistKeyName(k)
			if err != nil {
				return err
			}

			if tuple.GameVersion != version {
				return nil
			}

			out = append(out, tuple.Id)

			return nil
		})
	})
	if err != nil {
		return blizzardv2.ItemIds{}, err
	}

	return out, nil
}
