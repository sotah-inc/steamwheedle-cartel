package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewMiniRecipe(recipe Recipe) MiniRecipe {
	reagents := map[blizzardv2.ItemId]int{}
	for _, reagent := range recipe.BlizzardMeta.Reagents {
		reagents[reagent.Reagent.Id] = reagent.Quantity
	}

	return MiniRecipe{
		CraftedItemId:         recipe.BlizzardMeta.CraftedItem.Id,
		HordeCraftedItemId:    recipe.BlizzardMeta.HordeCraftedItem.Id,
		AllianceCraftedItemId: recipe.BlizzardMeta.AllianceCraftedItem.Id,
		Reagents:              reagents,
		CraftedQuantity:       recipe.BlizzardMeta.CraftedQuantity.Value,
	}
}

type MiniRecipe struct {
	CraftedItemId         blizzardv2.ItemId         `json:"crafted_item_id"`
	HordeCraftedItemId    blizzardv2.ItemId         `json:"horde_crafted_item_id"`
	AllianceCraftedItemId blizzardv2.ItemId         `json:"alliance_crafted_item_id"`
	Reagents              map[blizzardv2.ItemId]int `json:"reagents"`
	CraftedQuantity       float32                   `json:"crafted_quantity"`
}

func NewMiniRecipeMap(gzipEncoded []byte) (MiniRecipeMap, error) {
	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return MiniRecipeMap{}, err
	}

	out := MiniRecipeMap{}
	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return MiniRecipeMap{}, err
	}

	return out, nil
}

type MiniRecipeMap map[blizzardv2.RecipeId]MiniRecipe

func (mrMap MiniRecipeMap) EncodeForDelivery() ([]byte, error) {
	jsonEncoded, err := json.Marshal(mrMap)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}
