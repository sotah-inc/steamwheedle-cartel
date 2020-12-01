package sotah

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

// recipe-price-histories
type RecipeItemPriceHistory struct {
	Id           blizzardv2.ItemId `json:"id"`
	PriceHistory PriceHistory      `json:"price_history"`
}

type RecipePriceHistory struct {
	CraftedItem  RecipeItemPriceHistory `json:"crafted_item"`
	AllianceItem RecipeItemPriceHistory `json:"alliance_crafted_item"`
	HordeItem    RecipeItemPriceHistory `json:"horde_crafted_item"`
	Reagents     PriceHistory           `json:"reagents"`
}
