package stats

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (sBase Database) Stats() (sotah.AuctionStats, error) {
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
