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
				bkt, err := tx.CreateBucketIfNotExists(connectedRealmsBucketName())
				if err != nil {
					return err
				}

				for id, timestamps := range csStamps {
					k := connectedRealmsKeyName(blizzardv2.RegionVersionConnectedRealmTuple{
						RegionVersionTuple: blizzardv2.RegionVersionTuple{
							RegionTuple: blizzardv2.RegionTuple{RegionName: regionName},
							Version:     gameVersion,
						},
						ConnectedRealmId: id,
					})
					data := bkt.Get(k)
					if data == nil {
						return errors.New("could not find connected-realm by id")
					}

					realmComposite, err := sotah.NewRealmCompositeFromStorage(data)
					if err != nil {
						return err
					}

					realmComposite.StatusTimestamps = realmComposite.StatusTimestamps.Merge(timestamps)
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
