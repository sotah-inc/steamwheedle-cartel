package sotah

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

// recipe-price-histories
type RecipeItemPrices struct {
	Id     blizzardv2.ItemId `json:"id"`
	Prices Prices            `json:"prices"`
}

type RecipePrices struct {
	CraftedItemPrices  RecipeItemPrices `json:"crafted_item_prices"`
	AllianceItemPrices RecipeItemPrices `json:"alliance_crafted_item_prices"`
	HordeItemPrices    RecipeItemPrices `json:"horde_crafted_item_prices"`
	ReagentPrices      Prices           `json:"reagent_prices"`
}

type RecipePricesMap map[blizzardv2.RecipeId]RecipePrices
