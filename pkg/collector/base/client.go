package base

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type CollectAuctionsResult interface {
	Tuple() blizzardv2.LoadConnectedRealmTuple
	ItemIds() blizzardv2.ItemIds
}

type CollectAuctionsResults interface {
	ItemIds() blizzardv2.ItemIds
	RegionTimestamps() sotah.RegionTimestamps
	Tuples() blizzardv2.LoadConnectedRealmTuples
}

type Client interface {
	ResolveAuctions(tuples []blizzardv2.DownloadConnectedRealmTuple) chan blizzardv2.GetAuctionsJob
	GetTuples() []blizzardv2.DownloadConnectedRealmTuple
	ReceiveRegionTimestamps(timestamps sotah.RegionTimestamps)
	LakeClient() BaseLake.Client

	CollectAuctions() (CollectAuctionsResults, error)
}
