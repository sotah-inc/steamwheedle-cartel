package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) GetRegionTimestamps(
	version gameversion.GameVersion,
) (sotah.RegionTimestamps, error) {
	var out sotah.RegionTimestamps

	err := rBase.db.View(func(tx *bolt.Tx) error {
		baseBucket := tx.Bucket(baseBucketName())
		if baseBucket == nil {
			return errors.New("base-bucket does not exist")
		}

		return baseBucket.ForEach(func(baseBucketKey []byte, v []byte) error {
			name := regionNameFromKeyName(baseBucketKey)
			connectedRealmsBucket := tx.Bucket(connectedRealmsBucketName(version, name))
			if connectedRealmsBucket == nil {
				return errors.New("connected-realms bucket does not exist")
			}

			out[name] = map[blizzardv2.ConnectedRealmId]sotah.ConnectedRealmTimestamps{}

			return connectedRealmsBucket.ForEach(
				func(connectedRealmKey []byte, connectedRealmValue []byte) error {
					realmComposite, err := sotah.NewRealmCompositeFromStorage(connectedRealmValue)
					if err != nil {
						return err
					}

					out[name][realmComposite.ConnectedRealmResponse.Id] = realmComposite.ModificationDates

					return nil
				},
			)
		})
	})
	if err != nil {
		return sotah.RegionTimestamps{}, err
	}

	return out, nil
}
