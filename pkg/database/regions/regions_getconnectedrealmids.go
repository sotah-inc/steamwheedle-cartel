package regions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
)

func (rBase Database) GetConnectedRealmIds(
	version gameversion.GameVersion,
	name blizzardv2.RegionName,
) ([]blizzardv2.ConnectedRealmId, error) {
	var out []blizzardv2.ConnectedRealmId

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(connectedRealmsBucketName(version, name))
		if bkt == nil {
			return nil
		}

		return bkt.ForEach(func(k []byte, v []byte) error {
			id, err := connectedRealmIdFromKeyName(k)
			if err != nil {
				return err
			}

			out = append(out, id)

			return nil
		})
	})
	if err != nil {
		return []blizzardv2.ConnectedRealmId{}, err
	}

	return out, nil
}
