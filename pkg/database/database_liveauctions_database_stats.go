package database

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (ladBase LiveAuctionsDatabase) Stats() (sotah.MiniAuctionListStats, error) {
	maList, err := ladBase.GetMiniAuctionList()
	if err != nil {
		return sotah.MiniAuctionListStats{}, err
	}

	out := sotah.MiniAuctionListStats{
		MiniAuctionListGeneralStats: sotah.MiniAuctionListGeneralStats{
			TotalAuctions: maList.TotalAuctions(),
			TotalQuantity: maList.TotalQuantity(),
			TotalBuyout:   int(maList.TotalBuyout()),
		},
		ItemIds:    maList.ItemIds(),
		AuctionIds: maList.AuctionIds(),
	}

	return out, nil
}

func (ladBase LiveAuctionsDatabase) persistStats(currentTimestamp sotah.UnixTimestamp) error {
	stats, err := ladBase.Stats()
	if err != nil {
		return err
	}

	encodedData, err := stats.EncodeForStorage()
	if err != nil {
		return err
	}

	logging.WithFields(logrus.Fields{
		"db":           ladBase.db.Path(),
		"encoded-data": len(encodedData),
	}).Debug("Persisting mini-auction-stats via encoded-data")

	err = ladBase.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(liveAuctionsStatsBucketName())
		if err != nil {
			return err
		}

		if err := bkt.Put(liveAuctionsStatsKeyName(currentTimestamp), encodedData); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (ladBase LiveAuctionsDatabase) AuctionStats() (sotah.AuctionStats, error) {
	out := sotah.AuctionStats{}

	err := ladBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(liveAuctionsStatsBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			lastUpdated, err := unixTimestampFromLiveAuctionsStatsKeyName(k)
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
				NormalizeFunc:        normalizeLiveAuctionsStatsLastUpdated,
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
