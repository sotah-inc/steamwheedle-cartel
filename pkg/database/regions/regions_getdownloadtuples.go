package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/statuskinds"
)

func (rBase Database) GetDownloadTuples() ([]blizzardv2.DownloadConnectedRealmTuple, error) {
	var out []blizzardv2.DownloadConnectedRealmTuple

	err := rBase.db.View(func(tx *bolt.Tx) error {
		baseBucket := tx.Bucket(baseBucketName())
		if baseBucket == nil {
			return nil
		}

		connectedRealmsBucket := tx.Bucket(connectedRealmsBucketName())
		if connectedRealmsBucket == nil {
			return nil
		}

		return connectedRealmsBucket.ForEach(
			func(connectedRealmKey []byte, connectedRealmValue []byte) error {
				keyTuple, err := tupleFromConnectedRealmKeyName(connectedRealmKey)
				if err != nil {
					return err
				}

				regionValue := baseBucket.Get(baseKeyName(keyTuple.RegionName))
				if regionValue == nil {
					return errors.New("could not resolve region")
				}
				region, err := sotah.NewRegion(regionValue)
				if err != nil {
					return err
				}

				realmComposite, err := sotah.NewRealmCompositeFromStorage(connectedRealmValue)
				if err != nil {
					return err
				}

				lastModified := realmComposite.StatusTimestamps[statuskinds.Downloaded].Time()
				out = append(out, blizzardv2.DownloadConnectedRealmTuple{
					LoadConnectedRealmTuple: blizzardv2.LoadConnectedRealmTuple{
						RegionVersionConnectedRealmTuple: keyTuple,
						LastModified:                     lastModified,
					},
					RegionHostname: region.Hostname,
				})

				return nil
			},
		)
	})
	if err != nil {
		return []blizzardv2.DownloadConnectedRealmTuple{}, err
	}

	return out, nil
}
