package blizzardv2

import (
	"encoding/base64"
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewItemRecipesMap(base64Encoded string) (ItemRecipesMap, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return ItemRecipesMap{}, err
	}

	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return ItemRecipesMap{}, err
	}

	out := ItemRecipesMap{}
	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return ItemRecipesMap{}, err
	}

	return out, nil
}

type ItemRecipesMap map[ItemId]RecipeIds

func (irMap ItemRecipesMap) ItemIds() []ItemId {
	out := make([]ItemId, len(irMap))
	i := 0
	for id := range irMap {
		out[i] = id

		i += 1
	}

	return out
}

func (irMap ItemRecipesMap) Find(id ItemId) RecipeIds {
	found, ok := irMap[id]
	if !ok {
		return RecipeIds{}
	}

	return found
}

func (irMap ItemRecipesMap) Merge(input ItemRecipesMap) ItemRecipesMap {
	out := ItemRecipesMap{}
	for id, recipeIds := range irMap {
		out[id] = recipeIds
	}
	for id, providedRecipeIds := range input {
		foundRecipeIds := irMap.Find(id)
		out[id] = foundRecipeIds.Merge(providedRecipeIds)
	}

	return out
}

func (irMap ItemRecipesMap) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(irMap)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (irMap ItemRecipesMap) FilterBlank() ItemRecipesMap {
	out := ItemRecipesMap{}
	for itemId, recipeIds := range irMap {
		if len(recipeIds) == 0 {
			continue
		}

		out[itemId] = recipeIds
	}

	return out
}

func (irMap ItemRecipesMap) ToRecipesItemMap() map[RecipeId]ItemId {
	out := map[RecipeId]ItemId{}
	for itemId, recipeIds := range irMap {
		for _, recipeId := range recipeIds {
			out[recipeId] = itemId
		}
	}

	return out
}

func (irMap ItemRecipesMap) RecipeIds() RecipeIds {
	outMap := map[RecipeId]struct{}{}
	for _, recipeIds := range irMap {
		for _, recipeId := range recipeIds {
			outMap[recipeId] = struct{}{}
		}
	}

	out := make(RecipeIds, len(outMap))
	i := 0
	for recipeId := range outMap {
		out[i] = recipeId

		i += 1
	}

	return out
}
