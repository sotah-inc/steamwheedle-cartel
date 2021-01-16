package stats

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (sBase Database) PersistEncodedStats(
	currentTimestamp sotah.UnixTimestamp,
	encodedData []byte,
) error {
	logging.WithFields(logrus.Fields{
		"db":           sBase.db.Path(),
		"encoded-data": len(encodedData),
	}).Debug("persisting stats with encoded-data")

	err := sBase.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(baseBucketName())
		if err != nil {
			return err
		}

		if err := bkt.Put(baseKeyName(currentTimestamp), encodedData); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
