package base

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type GetEncodedAuctionsByTuplesJob interface {
	Err() error
	Tuple() blizzardv2.RegionConnectedRealmTuple
	EncodedAuctions() []byte

	ToLogrusFields() logrus.Fields
}

type Client interface {
	GetEncodedAuctionsByTuples(tuples []blizzardv2.RegionConnectedRealmTuple) chan GetEncodedAuctionsByTuplesJob
}
