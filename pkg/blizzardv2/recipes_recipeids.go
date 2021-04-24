package blizzardv2

import (
	"encoding/json"
)

func NewRecipeIdsMap(input RecipeIds) RecipeIdsMap {
	inputMap := map[RecipeId]struct{}{}
	for _, id := range input {
		inputMap[id] = struct{}{}
	}

	return inputMap
}

type RecipeIdsMap map[RecipeId]struct{}

func (idsMap RecipeIdsMap) ToIds() RecipeIds {
	out := make([]RecipeId, len(idsMap))
	i := 0
	for id := range idsMap {
		out[i] = id

		i += 1
	}

	return out
}

func NewRecipeIdsFromJson(jsonEncoded []byte) (RecipeIds, error) {
	out := RecipeIds{}

	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return RecipeIds{}, err
	}

	return out, nil
}

type RecipeIds []RecipeId

func (ids RecipeIds) Append(input RecipeIds) RecipeIds {
	out := make(RecipeIds, len(ids)+len(input))
	i := 0
	for _, id := range ids {
		out[i] = id

		i += 1
	}

	for _, id := range input {
		out[i] = id

		i += 1
	}

	return out
}

func (ids RecipeIds) IsZero() bool {
	return len(ids) == 0
}

func (ids RecipeIds) Subtract(input RecipeIds) RecipeIds {
	inputMap := NewRecipeIdsMap(input)

	out := RecipeIds{}
	for _, id := range ids {
		if _, ok := inputMap[id]; ok {
			continue
		}

		out = append(out, id)
	}

	return out
}

func (ids RecipeIds) Merge(input RecipeIds) RecipeIds {
	inputMap := NewRecipeIdsMap(ids)
	for _, id := range input {
		inputMap[id] = struct{}{}
	}

	return inputMap.ToIds()
}

func (ids RecipeIds) EncodeForStorage() ([]byte, error) {
	return json.Marshal(ids)
}
