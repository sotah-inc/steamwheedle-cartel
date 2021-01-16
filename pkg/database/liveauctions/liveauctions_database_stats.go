package liveauctions

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (ladBase Database) persistStats(currentTimestamp sotah.UnixTimestamp) error {
	maList, err := ladBase.GetMiniAuctionList()
	if err != nil {
		return err
	}

	encodedData, err := sotah.NewMiniAuctionListStatsFromMiniAuctionList(maList).EncodeForStorage()
	if err != nil {
		return err
	}

	logging.WithFields(logrus.Fields{
		"db":           ladBase.db.Path(),
		"encoded-data": len(encodedData),
	}).Debug("persisting mini-auction-stats via encoded-data")

	err = ladBase.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(statsBucketName())
		if err != nil {
			return err
		}

		if err := bkt.Put(statsKeyName(currentTimestamp), encodedData); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (ladBase Database) AuctionStats() (sotah.AuctionStats, error) {
	out := sotah.AuctionStats{}

	err := ladBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(statsBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			lastUpdated, err := unixTimestampFromStatsKeyName(k)
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
				NormalizeFunc:        normalizeStatsLastUpdated,
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
