package base

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type WriteAuctionsWithTuplesInJob interface {
	Tuple() blizzardv2.RegionVersionConnectedRealmTuple
	Auctions() sotah.MiniAuctionList
}

type WriteAuctionsWithTuplesOutJob interface {
	Err() error
	Tuple() blizzardv2.RegionVersionConnectedRealmTuple

	ToLogrusFields() logrus.Fields
}
