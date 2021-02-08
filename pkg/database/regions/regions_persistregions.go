package regions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) PersistRegions(regions sotah.RegionList) error {
	return rBase.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(baseBucketName())
		if err != nil {
			return err
		}

		for _, region := range regions {
			encodedRegion, err := region.EncodeForStorage()
			if err != nil {
				return err
			}

			if err := bkt.Put(baseKeyName(region.Name), encodedRegion); err != nil {
				return err
			}
		}

		return nil
	})
}
