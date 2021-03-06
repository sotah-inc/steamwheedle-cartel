package stats

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (sBase Database) PruneStats(retentionLimit sotah.UnixTimestamp) error {
	timestamps, err := sBase.getStatsTimestamps()
	if err != nil {
		return err
	}

	expiredTimestamps := timestamps.Before(retentionLimit)

	logging.WithFields(logrus.Fields{
		"db":         sBase.db.Path(),
		"timestamps": len(expiredTimestamps),
	}).Debug("pruning timestamps")

	err = sBase.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(baseBucketName())
		if err != nil {
			return err
		}

		for _, expiredTimestamp := range expiredTimestamps {
			if err := bkt.Delete(baseKeyName(expiredTimestamp)); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (sBase Database) getStatsTimestamps() (sotah.UnixTimestamps, error) {
	var out sotah.UnixTimestamps
	err := sBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			parsedKey, err := unixTimestampFromBaseKeyName(k)
			if err != nil {
				return err
			}

			out = append(out, parsedKey)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return []sotah.UnixTimestamp{}, err
	}

	return out, nil
}
