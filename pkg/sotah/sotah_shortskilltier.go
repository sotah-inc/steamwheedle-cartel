package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
)

func NewShortSkillTierCategoryRecipes(
	recipes []blizzardv2.SkillTierCategoryRecipe,
	providedLocale locale.Locale,
	providedRecipes map[blizzardv2.RecipeId]Recipe,
) []ShortSkillTierCategoryRecipe {
	out := make([]ShortSkillTierCategoryRecipe, len(recipes))
	for i, recipe := range recipes {
		foundRecipe := func() Recipe {
			foundRecipe, ok := providedRecipes[recipe.Id]
			if !ok {
				return Recipe{}
			}

			return foundRecipe
		}()

		out[i] = ShortSkillTierCategoryRecipe{
			Id:      recipe.Id,
			Name:    recipe.Name.FindOr(providedLocale, ""),
			IconUrl: foundRecipe.SotahMeta.IconUrl,
			Rank:    foundRecipe.BlizzardMeta.Rank,
		}
	}

	return out
}

type ShortSkillTierCategoryRecipe struct {
	Id      blizzardv2.RecipeId `json:"id"`
	Name    string              `json:"name"`
	IconUrl string              `json:"icon_url"`
	Rank    int                 `json:"rank"`
}

func NewShortSkillTierCategories(
	categories []blizzardv2.SkillTierCategory,
	providedLocale locale.Locale,
	providedRecipes map[blizzardv2.RecipeId]Recipe,
) []ShortSkillTierCategory {
	out := make([]ShortSkillTierCategory, len(categories))
	for i, category := range categories {
		out[i] = ShortSkillTierCategory{
			Name:    category.Name.FindOr(providedLocale, ""),
			Recipes: NewShortSkillTierCategoryRecipes(category.Recipes, providedLocale, providedRecipes),
		}
	}

	return out
}

type ShortSkillTierCategory struct {
	Name    string                         `json:"name"`
	Recipes []ShortSkillTierCategoryRecipe `json:"recipes"`
}

func NewShortSkillTier(
	skillTier SkillTier,
	providedLocale locale.Locale,
	providedRecipes map[blizzardv2.RecipeId]Recipe,
) ShortSkillTier {
	return ShortSkillTier{
		Id:   skillTier.BlizzardMeta.Id,
		Name: skillTier.BlizzardMeta.Name.FindOr(providedLocale, ""),
		Categories: NewShortSkillTierCategories(
			skillTier.BlizzardMeta.Categories,
			providedLocale,
			providedRecipes,
		),
	}
}

type ShortSkillTier struct {
	Id                blizzardv2.SkillTierId   `json:"id"`
	Name              string                   `json:"name"`
	MinimumSkillLevel int                      `json:"minimum_skill_level"`
	MaximumSkillLevel int                      `json:"maximum_skill_level"`
	Categories        []ShortSkillTierCategory `json:"categories"`
}
