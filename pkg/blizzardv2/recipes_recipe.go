package blizzardv2

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

type RecipeId int

type Recipe struct {
	LinksBase
	Id   RecipeId       `json:"id"`
	Name locale.Mapping `json:"name"`
}
