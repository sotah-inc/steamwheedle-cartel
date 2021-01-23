package stats

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBases RegionDatabases) TotalStats(
	names []blizzardv2.RegionName,
) (sotah.AuctionStats, error) {
	out := sotah.AuctionStats{}

	for _, name := range names {
		rBase, err := rBases.GetRegionDatabase(name)
		if err != nil {
			return sotah.AuctionStats{}, err
		}

		stats, err := rBase.Stats()
		if err != nil {
			return sotah.AuctionStats{}, err
		}

		out = out.Append(stats)
	}

	return out, nil
}
