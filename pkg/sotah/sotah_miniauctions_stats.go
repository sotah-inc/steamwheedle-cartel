package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func NewMiniAuctionListStats(jsonEncoded []byte) (MiniAuctionListStats, error) {
	var jsonDecoded MiniAuctionListStats
	if err := json.Unmarshal(jsonEncoded, &jsonDecoded); err != nil {
		return MiniAuctionListStats{}, err
	}

	return jsonDecoded, nil
}

func NewMiniAuctionListStatsFromMiniAuctionList(maList MiniAuctionList) MiniAuctionListStats {
	return MiniAuctionListStats{
		MiniAuctionListGeneralStats: MiniAuctionListGeneralStats{
			TotalAuctions: maList.TotalAuctions(),
			TotalQuantity: maList.TotalQuantity(),
			TotalBuyout:   int(maList.TotalBuyout()),
		},
		ItemIds:    maList.ItemIds(),
		AuctionIds: maList.AuctionIds(),
	}
}

type AuctionStats map[UnixTimestamp]MiniAuctionListGeneralStats

func (s AuctionStats) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(s)
}

type AuctionStatsSetOptions struct {
	LastUpdatedTimestamp UnixTimestamp
	Stats                MiniAuctionListStats
	NormalizeFunc        func(timestamp UnixTimestamp) UnixTimestamp
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

func (s MiniAuctionListGeneralStats) Add(
	v MiniAuctionListGeneralStats,
) MiniAuctionListGeneralStats {
	s.TotalQuantity += v.TotalQuantity
	s.TotalBuyout += v.TotalBuyout
	s.TotalAuctions += v.TotalAuctions

	return s
}

func (s MiniAuctionListGeneralStats) EncodeForStorage() ([]byte, error) {
	return json.Marshal(s)
}

type MiniAuctionListStats struct {
	MiniAuctionListGeneralStats
	ItemIds    []blizzardv2.ItemId
	AuctionIds []blizzardv2.AuctionId
}

func (s MiniAuctionListStats) EncodeForStorage() ([]byte, error) {
	return json.Marshal(s)
}
