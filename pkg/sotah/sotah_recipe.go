package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type RecipeMeta struct{}

func NewRecipe(gzipEncoded []byte) (Recipe, error) {
	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return Recipe{}, err
	}

	out := Recipe{}
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return Recipe{}, err
	}

	return out, nil
}

type Recipe struct {
	BlizzardMeta blizzardv2.RecipeResponse `json:"blizzard_meta"`
	SotahMeta    RecipeMeta                `json:"sotah_meta"`
}

func (recipe Recipe) EncodeForStorage() ([]byte, error) {
	jsonEncoded, err := json.Marshal(recipe)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}

func NewRecipeIntakeRequest(gzipEncoded []byte) (RecipesIntakeRequest, error) {
	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return RecipesIntakeRequest{}, err
	}

	out := &RecipesIntakeRequest{}
	if err := json.Unmarshal(jsonEncoded, out); err != nil {
		return RecipesIntakeRequest{}, err
	}

	return *out, nil
}

type RecipesIntakeRequest struct {
	RecipeIds []blizzardv2.RecipeId `json:"recipe_ids"`
}

func (req RecipesIntakeRequest) EncodeForDelivery() ([]byte, error) {
	jsonEncoded, err := json.Marshal(req)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}
