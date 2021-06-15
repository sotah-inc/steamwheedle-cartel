package regions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBase Database) GetRealmTimestamps(
	tuple blizzardv2.RegionVersionTuple,
) (sotah.RealmStatusTimestamps, error) {
	var out sotah.RealmStatusTimestamps

	err := rBase.db.View(func(tx *bolt.Tx) error {
		connectedRealmsBucket := tx.Bucket(connectedRealmsBucketName(tuple))
		if connectedRealmsBucket == nil {
			return errors.New("connected-realms bucket does not exist")
		}

		return connectedRealmsBucket.ForEach(
			func(connectedRealmKey []byte, connectedRealmValue []byte) error {
				realmComposite, err := sotah.NewRealmCompositeFromStorage(connectedRealmValue)
				if err != nil {
					return err
				}

				out[realmComposite.ConnectedRealmResponse.Id] = realmComposite.StatusTimestamps

				return nil
			},
		)
	})
	if err != nil {
		return sotah.RealmStatusTimestamps{}, err
	}

	return out, nil
}
