package regions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
)

type PersistConnectedRealmsInJob struct {
	Id   blizzardv2.ConnectedRealmId
	Data []byte
}

func (rBase Database) PersistConnectedRealms(
	version gameversion.GameVersion,
	regionName blizzardv2.RegionName,
	in chan PersistConnectedRealmsInJob,
) error {
	return rBase.db.Batch(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(connectedRealmsBucketName(version, regionName))
		if err != nil {
			return err
		}

		for job := range in {
			if err := bkt.Put(connectedRealmsKeyName(job.Id), job.Data); err != nil {
				return err
			}
		}

		return nil
	})
}
