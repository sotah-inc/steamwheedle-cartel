package regions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (rBase Database) GetRegionNames() ([]blizzardv2.RegionName, error) {
	var out []blizzardv2.RegionName

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return nil
		}

		return bkt.ForEach(func(k []byte, v []byte) error {
			name := regionNameFromKeyName(k)

			out = append(out, name)

			return nil
		})
	})
	if err != nil {
		return []blizzardv2.RegionName{}, err
	}

	return out, nil
}
