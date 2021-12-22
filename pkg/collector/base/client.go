package base

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type CollectAuctionsResults struct {
	VersionItems            blizzardv2.VersionItemsMap
	RegionVersionTimestamps sotah.RegionVersionTimestamps
	Tuples                  blizzardv2.LoadConnectedRealmTuples
}

type Client interface {
	Collect() error
	CollectAuctions() (CollectAuctionsResults, error)
}
