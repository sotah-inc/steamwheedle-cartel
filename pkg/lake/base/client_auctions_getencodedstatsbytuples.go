package base

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type GetEncodedStatsByTuplesJob interface {
	Err() error
	Tuple() blizzardv2.LoadConnectedRealmTuple
	EncodedStats() []byte

	ToLogrusFields() logrus.Fields
}
