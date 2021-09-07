package regions

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type PersistConnectedRealmsInJob struct {
	Id   blizzardv2.ConnectedRealmId
	Data []byte
}

func (rBase Database) PersistConnectedRealms(
	tuple blizzardv2.RegionVersionTuple,
	in chan PersistConnectedRealmsInJob,
) error {
	return rBase.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(connectedRealmsBucketName())
		if err != nil {
			return err
		}

		for job := range in {
			logging.WithFields(logrus.Fields{
				"tuple": tuple.String(),
				"realm": job.Id,
			}).Info("persisting")

			k := connectedRealmsKeyName(blizzardv2.RegionVersionConnectedRealmTuple{
				RegionVersionTuple: tuple,
				ConnectedRealmId:   job.Id,
			})
			if err := bkt.Put(k, job.Data); err != nil {
				return err
			}
		}

		return nil
	})
}
