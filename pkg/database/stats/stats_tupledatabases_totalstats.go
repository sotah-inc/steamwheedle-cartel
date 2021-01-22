package stats

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (tBases TupleDatabases) TotalStats(
	regionNames []blizzardv2.RegionName,
) (sotah.AuctionStats, error) {
	out := sotah.AuctionStats{}

	for _, regionName := range regionNames {
		auctionStats, err := tBases.RegionStats(regionName)
		if err != nil {
			return sotah.AuctionStats{}, err
		}

		out = out.Append(auctionStats)
	}

	return out, nil
}
