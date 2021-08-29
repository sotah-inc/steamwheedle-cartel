package regions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (rBase Database) GetConnectedRealmIds(
	tuple blizzardv2.RegionVersionTuple,
) ([]blizzardv2.ConnectedRealmId, error) {
	var out []blizzardv2.ConnectedRealmId

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(connectedRealmsBucketName())
		if bkt == nil {
			return nil
		}

		return bkt.ForEach(func(k []byte, v []byte) error {
			keyTuple, err := tupleFromConnectedRealmKeyName(k)
			if err != nil {
				return err
			}

			if !keyTuple.Equals(tuple) {
				return nil
			}

			out = append(out, keyTuple.ConnectedRealmId)

			return nil
		})
	})
	if err != nil {
		return []blizzardv2.ConnectedRealmId{}, err
	}

	return out, nil
}
