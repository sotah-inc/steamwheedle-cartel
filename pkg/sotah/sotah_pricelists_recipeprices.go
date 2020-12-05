package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewRecipeItemPrices(iPrices ItemPrices, id blizzardv2.ItemId) RecipeItemPrices {
	iPrice, ok := iPrices[id]
	if !ok {
		return RecipeItemPrices{
			Id: id,
			Prices: RecipeItemItemPrices{
				MinBuyoutPer:     0,
				MaxBuyoutPer:     0,
				AverageBuyoutPer: 0,
				MedianBuyoutPer:  0,
			},
		}
	}

	return RecipeItemPrices{
		Id: id,
		Prices: RecipeItemItemPrices{
			MinBuyoutPer:     iPrice.MinBuyoutPer,
			MaxBuyoutPer:     iPrice.MaxBuyoutPer,
			AverageBuyoutPer: iPrice.AverageBuyoutPer,
			MedianBuyoutPer:  iPrice.MedianBuyoutPer,
		},
	}
}

type RecipeItemPrices struct {
	Id     blizzardv2.ItemId    `json:"id"`
	Prices RecipeItemItemPrices `json:"prices"`
}

type RecipeItemItemPrices struct {
	MinBuyoutPer     float64 `json:"min_buyout_per"`
	MaxBuyoutPer     float64 `json:"max_buyout_per"`
	AverageBuyoutPer float64 `json:"average_buyout_per"`
	MedianBuyoutPer  float64 `json:"median_buyout_per"`
}

func NewRecipePrices(mRecipe MiniRecipe, iPrices ItemPrices) RecipePrices {
	totalReagentPrices := func() RecipeItemItemPrices {
		out := RecipeItemItemPrices{}
		for itemId, quantity := range mRecipe.Reagents {
			reagentPrices := NewRecipeItemPrices(iPrices, itemId).Prices

			if reagentPrices.MinBuyoutPer > 0 {
				out.MinBuyoutPer += reagentPrices.MinBuyoutPer * float64(quantity)
			}

			if reagentPrices.MaxBuyoutPer > 0 {
				out.MaxBuyoutPer += reagentPrices.MaxBuyoutPer * float64(quantity)
			}

			if reagentPrices.MedianBuyoutPer > 0 {
				out.MedianBuyoutPer += reagentPrices.MedianBuyoutPer * float64(quantity)
			}

			if reagentPrices.AverageBuyoutPer > 0 {
				out.AverageBuyoutPer += reagentPrices.AverageBuyoutPer * float64(quantity)
			}
		}

		return out
	}()

	return RecipePrices{
		CraftedItemPrices:  NewRecipeItemPrices(iPrices, mRecipe.CraftedItemId),
		AllianceItemPrices: NewRecipeItemPrices(iPrices, mRecipe.AllianceCraftedItemId),
		HordeItemPrices:    NewRecipeItemPrices(iPrices, mRecipe.HordeCraftedItemId),
		TotalReagentPrices: totalReagentPrices,
	}
}

func NewRecipePricesFromGzip(gzipEncoded []byte) (RecipePrices, error) {
	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return RecipePrices{}, err
	}

	out := RecipePrices{}
	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return RecipePrices{}, err
	}

	return out, nil
}

type RecipePrices struct {
	CraftedItemPrices  RecipeItemPrices     `json:"crafted_item_prices"`
	AllianceItemPrices RecipeItemPrices     `json:"alliance_crafted_item_prices"`
	HordeItemPrices    RecipeItemPrices     `json:"horde_crafted_item_prices"`
	TotalReagentPrices RecipeItemItemPrices `json:"total_reagent_prices"`
}

func (rPrices RecipePrices) EncodeForStorage() ([]byte, error) {
	jsonEncoded, err := json.Marshal(rPrices)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}

func NewRecipePricesMap(mRecipes MiniRecipes, iPrices ItemPrices) RecipePricesMap {
	out := RecipePricesMap{}
	for _, mRecipe := range mRecipes {
		out[mRecipe.Id] = NewRecipePrices(mRecipe, iPrices)
	}

	return out
}

type RecipePricesMap map[blizzardv2.RecipeId]RecipePrices
