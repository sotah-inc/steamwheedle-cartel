package stats

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (sBase Database) persistAuctionStats(
	stats sotah.MiniAuctionListStats,
	currentTimestamp sotah.UnixTimestamp,
) error {
	encodedData, err := stats.EncodeForStorage()
	if err != nil {
		return err
	}

	logging.WithFields(logrus.Fields{
		"db":           sBase.db.Path(),
		"encoded-data": len(encodedData),
	}).Debug("persisting mini-auction-stats via encoded-data")

	err = sBase.db.Update(func(tx *bolt.Tx) error {
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

func (sBase Database) AuctionStats() (sotah.AuctionStats, error) {
	out := sotah.AuctionStats{}

	err := sBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			lastUpdated, err := unixTimestampFromBaseKeyName(k)
			if err != nil {
				return err
			}

			stats, err := sotah.NewMiniAuctionListStats(v)
			if err != nil {
				return err
			}

			out = out.Set(sotah.AuctionStatsSetOptions{
				LastUpdatedTimestamp: lastUpdated,
				Stats:                stats,
				NormalizeFunc:        normalizeLastUpdated,
			})

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.AuctionStats{}, err
	}

	return out, nil
}
