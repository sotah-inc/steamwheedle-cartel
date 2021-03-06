package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (rBase Database) RegionExists(name blizzardv2.RegionName) (bool, error) {
	out := false

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return errors.New("base-bucket does not exist in regions database")
		}

		v := bkt.Get(baseKeyName(name))
		if v == nil {
			return nil
		}

		out = true

		return nil
	})
	if err != nil {
		return false, err
	}

	return out, nil
}
