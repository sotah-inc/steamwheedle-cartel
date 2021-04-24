package blizzardv2

import (
	"encoding/base64"
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewRecipeSubjectMap(base64Encoded string) (RecipeSubjectMap, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return RecipeSubjectMap{}, err
	}

	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return RecipeSubjectMap{}, err
	}

	out := RecipeSubjectMap{}
	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return RecipeSubjectMap{}, err
	}

	return out, nil
}

type RecipeSubjectMap map[RecipeId]string

func (rdMap RecipeSubjectMap) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(rdMap)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	base64Encoded := base64.StdEncoding.EncodeToString(gzipEncoded)

	return base64Encoded, nil
}
