package tokens

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (tBase Database) GetRegionHistory(tuple blizzardv2.RegionTuple) (TokenHistory, error) {
	out := TokenHistory{}

	err := tBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName(tuple.RegionName))
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			lastUpdated, err := lastUpdatedFromBaseKeyName(k)
			if err != nil {
				return err
			}

			out[lastUpdated] = priceFromTokenValue(v)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return TokenHistory{}, err
	}

	return out, nil
}
