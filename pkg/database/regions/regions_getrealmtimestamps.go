package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) GetRealmTimestamps(
	version gameversion.GameVersion,
	name blizzardv2.RegionName,
) (sotah.RealmTimestamps, error) {
	var out sotah.RealmTimestamps

	err := rBase.db.View(func(tx *bolt.Tx) error {
		connectedRealmsBucket := tx.Bucket(connectedRealmsBucketName(version, name))
		if connectedRealmsBucket == nil {
			return errors.New("connected-realms bucket does not exist")
		}

		return connectedRealmsBucket.ForEach(
			func(connectedRealmKey []byte, connectedRealmValue []byte) error {
				realmComposite, err := sotah.NewRealmCompositeFromStorage(connectedRealmValue)
				if err != nil {
					return err
				}

				out[realmComposite.ConnectedRealmResponse.Id] = realmComposite.ModificationDates

				return nil
			},
		)
	})
	if err != nil {
		return sotah.RealmTimestamps{}, err
	}

	return out, nil
}
