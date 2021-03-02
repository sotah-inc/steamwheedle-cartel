package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewRecipeIdNameMap(data []byte) (RecipeIdNameMap, error) {
	gzipDecoded, err := util.GzipDecode(data)
	if err != nil {
		return RecipeIdNameMap{}, err
	}

	var out RecipeIdNameMap
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return RecipeIdNameMap{}, err
	}

	return out, nil
}

type RecipeIdNameMap map[blizzardv2.RecipeId]locale.Mapping

func (idNameMap RecipeIdNameMap) EncodeForDelivery() ([]byte, error) {
	jsonEncodedData, err := json.Marshal(idNameMap)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncodedData)
}

func (idNameMap RecipeIdNameMap) RecipeIds() []blizzardv2.RecipeId {
	out := make([]blizzardv2.RecipeId, len(idNameMap))
	i := 0
	for id := range idNameMap {
		out[i] = id

		i++
	}

	return out
}
