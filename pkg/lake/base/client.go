package base

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type Client interface {
	GetEncodedAuctionsByTuples(tuples blizzardv2.RegionConnectedRealmTuples) chan GetEncodedAuctionsByTuplesJob
	GetEncodedPricelistHistoryByTuples(
		tuples blizzardv2.LoadConnectedRealmTuples,
	) chan GetEncodedPricelistHistoryByTuplesJob
	WriteAuctionsWithTuples(in chan WriteAuctionsWithTuplesInJob) chan WriteAuctionsWithTuplesOutJob
	NewWriteAuctionsWithTuplesInJob(
		tuple blizzardv2.RegionConnectedRealmTuple,
		auctions sotah.MiniAuctionList,
	) WriteAuctionsWithTuplesInJob
}
