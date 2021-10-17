package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
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
			keyTuple, err := tupleFromConnectedRealmKeyName(k)
			if err != nil {
				return err
			}

			if !keyTuple.RegionVersionTuple.Equals(tuple.RegionVersionTuple) {
				logging.WithFields(logrus.Fields{
					"key-tuple": keyTuple.RegionVersionTuple.String(),
					"tuple":     tuple.RegionVersionTuple.String(),
				}).Debug("key-tuple was not equal to provided tuple")

				return nil
			}

			logging.WithFields(logrus.Fields{
				"key-tuple": keyTuple.String(),
				"tuple":     tuple.String(),
			}).Debug("key-tuple was equal to provided tuple")

			realmComposite, err := sotah.NewRealmCompositeFromStorage(v)
			if err != nil {
				return err
			}

			for _, realm := range realmComposite.ConnectedRealmResponse.Realms {
				if realm.Slug != tuple.RealmSlug {
					logging.WithFields(logrus.Fields{
						"realm-slug":       realm.Slug,
						"tuple-realm-slug": tuple.RealmSlug,
					}).Debug("realm-slug was not equal to provided tuple")

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
