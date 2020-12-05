package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func NewMiniRecipe(recipe Recipe) MiniRecipe {
	reagents := map[blizzardv2.ItemId]int{}
	for _, reagent := range recipe.BlizzardMeta.Reagents {
		reagents[reagent.Reagent.Id] = reagent.Quantity
	}

	return MiniRecipe{
		Id:                    recipe.BlizzardMeta.Id,
		CraftedItemId:         recipe.BlizzardMeta.CraftedItem.Id,
		HordeCraftedItemId:    recipe.BlizzardMeta.HordeCraftedItem.Id,
		AllianceCraftedItemId: recipe.BlizzardMeta.AllianceCraftedItem.Id,
		Reagents:              reagents,
		CraftedQuantity:       recipe.BlizzardMeta.CraftedQuantity.Value,
	}
}

type MiniRecipe struct {
	Id                    blizzardv2.RecipeId       `json:"id"`
	CraftedItemId         blizzardv2.ItemId         `json:"crafted_item_id"`
	HordeCraftedItemId    blizzardv2.ItemId         `json:"horde_crafted_item_id"`
	AllianceCraftedItemId blizzardv2.ItemId         `json:"alliance_crafted_item_id"`
	Reagents              map[blizzardv2.ItemId]int `json:"reagents"`
	CraftedQuantity       float32                   `json:"crafted_quantity"`
}

type MiniRecipes []MiniRecipe
