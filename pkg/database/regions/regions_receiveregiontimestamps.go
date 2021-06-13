package regions

import (
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) ReceiveRegionTimestamps(
	rvtStamps sotah.RegionVersionTimestamps,
) error {
	return rBase.db.Batch(func(tx *bolt.Tx) error {
		for regionName, vrStamps := range rvtStamps {
			for gameVersion, csStamps := range vrStamps {
				bkt, err := tx.CreateBucketIfNotExists(connectedRealmsBucketName(blizzardv2.RegionVersionTuple{
					RegionTuple: blizzardv2.RegionTuple{
						RegionName: regionName,
					},
					Version: gameVersion,
				}))
				if err != nil {
					return err
				}

				for id, timestamps := range csStamps {
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
		}

		return nil
	})
}
