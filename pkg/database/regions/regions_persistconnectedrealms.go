package regions

import (
	"errors"

	"github.com/boltdb/bolt"
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
	return rBase.db.Batch(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(connectedRealmsBucketName())
		if err != nil {
			return err
		}

		for job := range in {
			logging.WithField("id", job.Id).Info("persisting")

			if true {
				return errors.New("POOOOOOOOOP")
			}

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
