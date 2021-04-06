package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
)

type ShortRecipeReagent struct {
	Reagent  ShortRecipeItem `json:"reagent"`
	Quantity int             `json:"quantity"`
}

func NewShortRecipeReagents(
	reagents []blizzardv2.RecipeReagent,
	providedLocale locale.Locale,
) []ShortRecipeReagent {
	out := make([]ShortRecipeReagent, len(reagents))
	for i, reagent := range reagents {
		out[i] = ShortRecipeReagent{
			Reagent:  NewShortRecipeItem(reagent.Reagent, providedLocale),
			Quantity: reagent.Quantity,
		}
	}

	return out
}

func NewShortRecipeItem(
	item blizzardv2.RecipeItem,
	providedLocale locale.Locale,
) ShortRecipeItem {
	return ShortRecipeItem{
		Id:   item.Id,
		Name: item.Name.FindOr(providedLocale, ""),
	}
}

type ShortRecipeItem struct {
	Id   blizzardv2.ItemId `json:"id"`
	Name string            `json:"name"`
}

func NewShortRecipe(recipe Recipe, providedLocale locale.Locale) ShortRecipe {
	return ShortRecipe{
		Id:           recipe.BlizzardMeta.Id,
		ProfessionId: recipe.SotahMeta.ProfessionId,
		SkillTierId:  recipe.SotahMeta.SkillTierId,
		Name:         recipe.BlizzardMeta.Name.FindOr(providedLocale, ""),
		Description:  recipe.BlizzardMeta.Description.FindOr(providedLocale, ""),
		CraftedItem:  NewShortRecipeItem(recipe.BlizzardMeta.CraftedItem, providedLocale),
		AllianceCraftedItem: NewShortRecipeItem(
			recipe.BlizzardMeta.AllianceCraftedItem,
			providedLocale,
		),
		HordeCraftedItem: NewShortRecipeItem(
			recipe.BlizzardMeta.HordeCraftedItem,
			providedLocale,
		),
		SupplementalCraftedItemId: recipe.SotahMeta.SupplementalCraftedItemId,
		Reagents:                  NewShortRecipeReagents(recipe.BlizzardMeta.Reagents, providedLocale),
		Rank:                      recipe.BlizzardMeta.Rank,
		CraftedQuantity:           recipe.BlizzardMeta.CraftedQuantity.Value,
		IconUrl:                   recipe.SotahMeta.IconUrl,
	}
}

type ShortRecipe struct {
	Id                        blizzardv2.RecipeId     `json:"id"`
	ProfessionId              blizzardv2.ProfessionId `json:"profession_id"`
	SkillTierId               blizzardv2.SkillTierId  `json:"skilltier_id"`
	Name                      string                  `json:"name"`
	Description               string                  `json:"description"`
	CraftedItem               ShortRecipeItem         `json:"crafted_item"`
	AllianceCraftedItem       ShortRecipeItem         `json:"alliance_crafted_item"`
	HordeCraftedItem          ShortRecipeItem         `json:"horde_crafted_item"`
	SupplementalCraftedItemId blizzardv2.ItemId       `json:"supplemental_crafted_item_id"`
	Reagents                  []ShortRecipeReagent    `json:"reagents"`
	Rank                      int                     `json:"rank"`
	CraftedQuantity           float32                 `json:"crafted_quantity"`
	IconUrl                   string                  `json:"icon_url"`
}
