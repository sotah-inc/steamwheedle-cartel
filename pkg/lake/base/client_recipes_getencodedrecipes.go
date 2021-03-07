package base

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type GetEncodedRecipeJob interface {
	Err() error
	Id() blizzardv2.RecipeId
	CraftedItemIds() []blizzardv2.ItemId
	EncodedRecipe() []byte
	EncodedNormalizedName() []byte
	ToLogrusFields() logrus.Fields
}
