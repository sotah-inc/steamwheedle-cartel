package database

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (ladBase LiveAuctionsDatabase) pruneStats(retentionLimit sotah.UnixTimestamp) error {
	timestamps, err := ladBase.getStatsTimestamps()
	if err != nil {
		return err
	}

	expiredTimestamps := timestamps.Before(retentionLimit)

	logging.WithFields(logrus.Fields{
		"db":         ladBase.db.Path(),
		"timestamps": len(expiredTimestamps),
	}).Debug("pruning timestamps")

	err = ladBase.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(liveAuctionsStatsBucketName())
		if err != nil {
			return err
		}

		for _, expiredTimestamp := range expiredTimestamps {
			if err := bkt.Delete(liveAuctionsStatsKeyName(expiredTimestamp)); err != nil {
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

func (ladBase LiveAuctionsDatabase) getStatsTimestamps() (sotah.UnixTimestamps, error) {
	var out sotah.UnixTimestamps
	err := ladBase.db.View(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(liveAuctionsStatsBucketName())
		if err != nil {
			return err
		}

		err = bkt.ForEach(func(k, v []byte) error {
			parsedKey, err := unixTimestampFromLiveAuctionsStatsKeyName(k)
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
