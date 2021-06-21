package regions

import (
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"github.com/boltdb/bolt"
)

func (rBase Database) ConnectedRealmExists(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
) (bool, error) {
	out := false

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(connectedRealmsBucketName(tuple.RegionVersionTuple))
		if bkt == nil {
			return errors.New("connected-realms-bucket does not exist in regions database")
		}

		v := bkt.Get(connectedRealmsKeyName(tuple.ConnectedRealmId))
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
