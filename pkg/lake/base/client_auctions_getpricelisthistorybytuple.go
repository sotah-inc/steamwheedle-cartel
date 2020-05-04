package base

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type GetEncodedPricelistHistoryByTuplesJob interface {
	Err() error
	Tuple() blizzardv2.LoadConnectedRealmTuple
	EncodedPricelistHistory() map[blizzardv2.ItemId][]byte

	ToLogrusFields() logrus.Fields
}
