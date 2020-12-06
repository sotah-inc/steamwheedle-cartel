package base

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type GetEncodedRecipePricesByTuplesJob interface {
	Err() error
	Tuple() blizzardv2.LoadConnectedRealmTuple
	EncodedRecipePrices() map[blizzardv2.RecipeId][]byte

	ToLogrusFields() logrus.Fields
}
