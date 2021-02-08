package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (rBase Database) ConnectedRealmExists(
	name blizzardv2.RegionName,
	id blizzardv2.ConnectedRealmId,
) (bool, error) {
	out := false

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(connectedRealmsBucketName(name))
		if bkt == nil {
			return errors.New("connected-realms-bucket does not exist in regions database")
		}

		v := bkt.Get(connectedRealmsKeyName(id))
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
