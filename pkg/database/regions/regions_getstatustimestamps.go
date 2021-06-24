package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) GetStatusTimestamps(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
) (sotah.StatusTimestamps, error) {
	out := sotah.StatusTimestamps{}

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(connectedRealmsBucketName())
		if bkt == nil {
			return nil
		}

		v := bkt.Get(connectedRealmsKeyName(tuple))
		if v == nil {
			return errors.New("could not resolve connected-realm")
		}

		realm, err := sotah.NewRealmCompositeFromStorage(v)
		if err != nil {
			return err
		}

		out = realm.StatusTimestamps

		return nil
	})
	if err != nil {
		return sotah.StatusTimestamps{}, err
	}

	return out, nil
}
