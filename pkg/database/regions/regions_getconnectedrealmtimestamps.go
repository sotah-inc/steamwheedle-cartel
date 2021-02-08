package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) GetConnectedRealmsTimestamps(
	name blizzardv2.RegionName,
	id blizzardv2.ConnectedRealmId,
) (sotah.ConnectedRealmTimestamps, error) {
	out := sotah.ConnectedRealmTimestamps{}

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(connectedRealmsBucketName(name))
		if bkt == nil {
			return nil
		}

		v := bkt.Get(connectedRealmsKeyName(id))
		if v == nil {
			return errors.New("could not resolve connected-realm")
		}

		realm, err := sotah.NewRealmCompositeFromStorage(v)
		if err != nil {
			return err
		}

		out = realm.ModificationDates

		return nil
	})
	if err != nil {
		return sotah.ConnectedRealmTimestamps{}, err
	}

	return out, nil
}
