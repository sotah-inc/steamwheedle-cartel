package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) RealmExists(
	name blizzardv2.RegionName,
	version gameversion.GameVersion,
	slug blizzardv2.RealmSlug,
) (bool, error) {
	out := false

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(connectedRealmsBucketName(name, version))
		if bkt == nil {
			return errors.New("connected-realms-bucket does not exist in regions database")
		}

		return bkt.ForEach(func(k []byte, v []byte) error {
			connectedRealm, err := sotah.NewRealmCompositeFromStorage(v)
			if err != nil {
				return err
			}

			for _, realm := range connectedRealm.ConnectedRealmResponse.Realms {
				if realm.Slug == slug {
					out = true

					return nil
				}
			}

			return nil
		})
	})
	if err != nil {
		return false, err
	}

	return out, nil
}
