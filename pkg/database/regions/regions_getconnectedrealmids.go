package regions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (rBase Database) GetConnectedRealmIds(
	name blizzardv2.RegionName,
) ([]blizzardv2.ConnectedRealmId, error) {
	var out []blizzardv2.ConnectedRealmId

	err := rBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(connectedRealmsBucketName(name))
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
