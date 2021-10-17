package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) GetConnectedRealmBySlugTuple(
	tuple blizzardv2.RegionVersionRealmTuple,
) (sotah.RealmComposite, error) {
	out := sotah.RealmComposite{}

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(connectedRealmsBucketName())
		if bkt == nil {
			logging.Error("connected-realms bucket was nil")

			return nil
		}

		return bkt.ForEach(func(k []byte, v []byte) error {
			if !out.IsZero() {
				return nil
			}

			keyTuple, err := tupleFromConnectedRealmKeyName(k)
			if err != nil {
				return err
			}

			if !keyTuple.RegionVersionTuple.Equals(tuple.RegionVersionTuple) {
				return nil
			}

			realmComposite, err := sotah.NewRealmCompositeFromStorage(v)
			if err != nil {
				return err
			}

			for _, realm := range realmComposite.ConnectedRealmResponse.Realms {
				if realm.Slug != tuple.RealmSlug {
					continue
				}

				out = realmComposite

				return nil
			}

			return errors.New("could not resolve connected-realm with slug")
		})
	})
	if err != nil {
		return sotah.RealmComposite{}, err
	}

	return out, nil
}
