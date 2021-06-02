package regions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) GetConnectedRealms(
	version gameversion.GameVersion,
	name blizzardv2.RegionName,
) (sotah.RealmComposites, error) {
	out := sotah.RealmComposites{}

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(connectedRealmsBucketName(version, name))
		if bkt == nil {
			return nil
		}

		return bkt.ForEach(func(k []byte, v []byte) error {
			realm, err := sotah.NewRealmCompositeFromStorage(v)
			if err != nil {
				return err
			}

			out = append(out, realm)

			return nil
		})
	})
	if err != nil {
		return sotah.RealmComposites{}, err
	}

	return out, nil
}
