package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) GetDownloadTuples(
	version gameversion.GameVersion,
) ([]blizzardv2.DownloadConnectedRealmTuple, error) {
	var out []blizzardv2.DownloadConnectedRealmTuple

	err := rBase.db.View(func(tx *bolt.Tx) error {
		baseBucket := tx.Bucket(baseBucketName())
		if baseBucket == nil {
			return errors.New("base-bucket does not exist")
		}

		return baseBucket.ForEach(func(baseBucketKey []byte, regionValue []byte) error {
			region, err := sotah.NewRegion(regionValue)
			if err != nil {
				return err
			}

			connectedRealmsBucket := tx.Bucket(connectedRealmsBucketName(blizzardv2.RegionVersionTuple{
				RegionTuple: blizzardv2.RegionTuple{RegionName: region.Name},
				Version:     version,
			}))
			if connectedRealmsBucket == nil {
				return errors.New("connected-realms bucket does not exist")
			}

			return connectedRealmsBucket.ForEach(
				func(connectedRealmKey []byte, connectedRealmValue []byte) error {
					realmComposite, err := sotah.NewRealmCompositeFromStorage(connectedRealmValue)
					if err != nil {
						return err
					}

					out = append(out, blizzardv2.DownloadConnectedRealmTuple{
						LoadConnectedRealmTuple: blizzardv2.LoadConnectedRealmTuple{
							RegionVersionConnectedRealmTuple: blizzardv2.RegionVersionConnectedRealmTuple{
								RegionVersionTuple: blizzardv2.RegionVersionTuple{
									RegionTuple: blizzardv2.RegionTuple{RegionName: region.Name},
									Version:     version,
								},
								ConnectedRealmId: realmComposite.ConnectedRealmResponse.Id,
							},
							LastModified: realmComposite.StatusTimestamps["downloaded"].Time(),
						},
						RegionHostname: region.Hostname,
					})

					return nil
				},
			)
		})
	})
	if err != nil {
		return []blizzardv2.DownloadConnectedRealmTuple{}, err
	}

	return out, nil
}
