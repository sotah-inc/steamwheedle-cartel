package stats

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (tBases TupleDatabases) RegionStats(regionName blizzardv2.RegionName) (sotah.AuctionStats, error) {
	shard, err := tBases.GetRegionShard(regionName)
	if err != nil {
		return sotah.AuctionStats{}, err
	}

	out := sotah.AuctionStats{}
	for _, db := range shard {
		auctionStats, err := db.Stats()
		if err != nil {
			return sotah.AuctionStats{}, err
		}

		out = out.Append(auctionStats)

	}

	return out, nil
}
