package database

import (
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewMiniAuctionListStats(gzipEncoded []byte) (MiniAuctionListStats, error) {
	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return MiniAuctionListStats{}, err
	}

	var jsonDecoded MiniAuctionListStats
	if err := json.Unmarshal(gzipDecoded, &jsonDecoded); err != nil {
		return MiniAuctionListStats{}, err
	}

	return jsonDecoded, nil
}

type AuctionStats map[int64]MiniAuctionListGeneralStats

func (s AuctionStats) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(s)
}

func (s AuctionStats) Set(lastUpdatedTimestamp int64, stats MiniAuctionListStats) AuctionStats {
	s[normalizeLiveAuctionsStatsLastUpdated(lastUpdatedTimestamp)] = stats.MiniAuctionListGeneralStats

	return s
}

func (s AuctionStats) Append(nextStats AuctionStats) AuctionStats {
	for k, v := range nextStats {
		next := func() MiniAuctionListGeneralStats {
			found, ok := s[k]
			if !ok {
				return v
			}

			return v.Add(found)
		}()

		s[k] = next
	}

	return s
}

type MiniAuctionListGeneralStats struct {
	TotalAuctions int `json:"total_auctions"`
	TotalQuantity int `json:"total_quantity"`
	TotalBuyout   int `json:"total_buyout"`
}

func (s MiniAuctionListGeneralStats) Add(v MiniAuctionListGeneralStats) MiniAuctionListGeneralStats {
	s.TotalQuantity += v.TotalQuantity
	s.TotalBuyout += v.TotalBuyout
	s.TotalAuctions += v.TotalAuctions

	return s
}

type MiniAuctionListStats struct {
	MiniAuctionListGeneralStats
	ItemIds    []blizzardv2.ItemId
	AuctionIds []blizzardv2.AuctionId
}

func (s MiniAuctionListStats) EncodeForStorage() ([]byte, error) {
	jsonEncoded, err := json.Marshal(s)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}

func (ladBase LiveAuctionsDatabase) Stats() (MiniAuctionListStats, error) {
	maList, err := ladBase.GetMiniAuctionList()
	if err != nil {
		return MiniAuctionListStats{}, err
	}

	out := MiniAuctionListStats{
		MiniAuctionListGeneralStats: MiniAuctionListGeneralStats{
			TotalAuctions: maList.TotalAuctions(),
			TotalQuantity: maList.TotalQuantity(),
			TotalBuyout:   int(maList.TotalBuyout()),
		},
		ItemIds:    maList.ItemIds(),
		AuctionIds: maList.AuctionIds(),
	}

	return out, nil
}

func (ladBase LiveAuctionsDatabase) persistStats(currentTime time.Time) error {
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

		if err := bkt.Put(liveAuctionsStatsKeyName(currentTime.Unix()), encodedData); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (ladBase LiveAuctionsDatabase) GetAuctionStats() (AuctionStats, error) {
	out := AuctionStats{}

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

			stats, err := NewMiniAuctionListStats(v)
			if err != nil {
				return err
			}

			out = out.Set(lastUpdated, stats)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return AuctionStats{}, err
	}

	return out, nil
}
