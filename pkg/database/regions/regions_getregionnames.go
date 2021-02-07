package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (rBase Database) GetRegionNames() ([]blizzardv2.RegionName, error) {
	var out []blizzardv2.RegionName

	err := rBase.db.View(func(tx *bolt.Tx) error {
		baseBucket := tx.Bucket(baseBucketName())
		if baseBucket == nil {
			return errors.New("base-bucket does not exist in regions database")
		}

		return baseBucket.ForEach(func(k []byte, v []byte) error {
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
