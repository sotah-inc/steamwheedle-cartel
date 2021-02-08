package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) GetRegion(name blizzardv2.RegionName) (sotah.Region, error) {
	out := sotah.Region{}

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return errors.New("base-bucket does not exist in regions database")
		}

		v := bkt.Get(baseKeyName(name))
		if v == nil {
			return errors.New("region not found")
		}

		var err error
		out, err = sotah.NewRegion(v)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.Region{}, err
	}

	return out, nil
}
