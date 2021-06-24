package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) GetRegionTimestamps(
	version gameversion.GameVersion,
) (sotah.RegionVersionTimestamps, error) {
	var out sotah.RegionVersionTimestamps

	err := rBase.db.View(func(tx *bolt.Tx) error {
		baseBucket := tx.Bucket(baseBucketName())
		if baseBucket == nil {
			return errors.New("base-bucket does not exist")
		}

		return baseBucket.ForEach(func(baseBucketKey []byte, v []byte) error {
			name := regionNameFromKeyName(baseBucketKey)
			connectedRealmsBucket := tx.Bucket(connectedRealmsBucketName())
			if connectedRealmsBucket == nil {
				return errors.New("connected-realms bucket does not exist")
			}

			out[name] = sotah.VersionRealmTimestamps{version: sotah.RealmStatusTimestamps{}}

			return connectedRealmsBucket.ForEach(
				func(connectedRealmKey []byte, connectedRealmValue []byte) error {
					keyTuple, err := tupleFromConnectedRealmKeyName(connectedRealmKey)
					if err != nil {
						return err
					}

					if keyTuple.RegionName != name || keyTuple.Version != version {
						return nil
					}

					realmComposite, err := sotah.NewRealmCompositeFromStorage(connectedRealmValue)
					if err != nil {
						return err
					}

					out[name][version][realmComposite.ConnectedRealmResponse.Id] = realmComposite.StatusTimestamps

					return nil
				},
			)
		})
	})
	if err != nil {
		return sotah.RegionVersionTimestamps{}, err
	}

	return out, nil
}
