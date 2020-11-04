package base

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type GetEncodedRecipeJob interface {
	Err() error
	Id() blizzardv2.RecipeId
	EncodedRecipe() []byte
	ToLogrusFields() logrus.Fields
}
