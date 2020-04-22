package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
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

type AuctionStats map[sotah.UnixTimestamp]MiniAuctionListGeneralStats

func (s AuctionStats) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(s)
}

type AuctionStatsSetOptions struct {
	LastUpdatedTimestamp sotah.UnixTimestamp
	Stats                MiniAuctionListStats
	NormalizeFunc        func(timestamp sotah.UnixTimestamp) sotah.UnixTimestamp
}

func (s AuctionStats) Set(opts AuctionStatsSetOptions) AuctionStats {
	s[opts.NormalizeFunc(opts.LastUpdatedTimestamp)] = opts.Stats.MiniAuctionListGeneralStats

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
