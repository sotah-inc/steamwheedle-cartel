package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) ReceiveRegionTimestamps(
	version gameversion.GameVersion,
	regionTimestamps sotah.RegionTimestamps,
) error {
	return rBase.db.Batch(func(tx *bolt.Tx) error {
		for regionName, connectedRealmTimestamps := range regionTimestamps {
			bkt, err := tx.CreateBucketIfNotExists(connectedRealmsBucketName(regionName, version))
			if err != nil {
				return err
			}

			for id, timestamps := range connectedRealmTimestamps {
				k := connectedRealmsKeyName(id)
				data := bkt.Get(k)
				if data == nil {
					return errors.New("could not find connected-realm by id")
				}

				realmComposite, err := sotah.NewRealmCompositeFromStorage(data)
				if err != nil {
					return err
				}

				realmComposite.ModificationDates = realmComposite.ModificationDates.Merge(timestamps)
				encoded, err := realmComposite.EncodeForStorage()
				if err != nil {
					return err
				}

				if err := bkt.Put(k, encoded); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
