package regions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) GetConnectedRealmByRealmSlug(
	version gameversion.GameVersion,
	name blizzardv2.RegionName,
	slug blizzardv2.RealmSlug,
) (sotah.RealmComposite, error) {
	out := sotah.RealmComposite{}

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(connectedRealmsBucketName(version, name))
		if bkt == nil {
			return nil
		}

		return bkt.ForEach(func(k []byte, v []byte) error {
			realmComposite, err := sotah.NewRealmCompositeFromStorage(v)
			if err != nil {
				return err
			}

			for _, realm := range realmComposite.ConnectedRealmResponse.Realms {
				if realm.Slug != slug {
					continue
				}

				out = realmComposite
			}

			return nil
		})
	})
	if err != nil {
		return sotah.RealmComposite{}, err
	}

	return out, nil
}
