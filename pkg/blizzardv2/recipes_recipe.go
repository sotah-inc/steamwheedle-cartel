package blizzardv2

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

type RecipeId int

type RecipeItem struct {
	Key  HrefReference  `json:"key"`
	Name locale.Mapping `json:"name"`
	Id   ItemId         `json:"id"`
}

func (item RecipeItem) IsZero() bool {
	return item.Id == 0
}

type RecipeReagent struct {
	Reagent  RecipeItem `json:"reagent"`
	Quantity int        `json:"quantity"`
}

type RecipeModifiedCraftingSlots struct {
	SlotType struct {
		Key HrefReference `json:"key"`
		Id  int           `json:"id"`
	} `json:"slot_type"`
	DisplayOrder int `json:"display_order"`
}

type Recipe struct {
	LinksBase
	Id          RecipeId       `json:"id"`
	Name        locale.Mapping `json:"name"`
	Description locale.Mapping `json:"description"`
	Media       struct {
		Key HrefReference `json:"key"`
		Id  RecipeId      `json:"id"`
	} `json:"media"`
	AllianceCraftedItem RecipeItem      `json:"alliance_crafted_item"`
	HordeCraftedItem    RecipeItem      `json:"horde_crafted_item"`
	Reagents            []RecipeReagent `json:"reagents"`
	CraftedQuantity     struct {
		Value float32 `json:"value"`
	} `json:"crafted_quantity"`
	ModifiedCraftingSlots []RecipeModifiedCraftingSlots `json:"modified_crafting_slots"`
}
