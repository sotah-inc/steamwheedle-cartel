package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
)

func NewShortSkillTierCategoryRecipes(
	recipes []blizzardv2.SkillTierCategoryRecipe,
	providedLocale locale.Locale,
	recipeIconUrls map[blizzardv2.RecipeId]string,
) []ShortSkillTierCategoryRecipe {
	out := make([]ShortSkillTierCategoryRecipe, len(recipes))
	for i, recipe := range recipes {
		iconUrl := func() string {
			foundIconUrl, ok := recipeIconUrls[recipe.Id]
			if !ok {
				return ""
			}

			return foundIconUrl
		}()

		out[i] = ShortSkillTierCategoryRecipe{
			Id:      recipe.Id,
			Name:    recipe.Name.FindOr(providedLocale, ""),
			IconUrl: iconUrl,
		}
	}

	return out
}

type ShortSkillTierCategoryRecipe struct {
	Id      blizzardv2.RecipeId `json:"id"`
	Name    string              `json:"name"`
	IconUrl string              `json:"icon_url"`
}

func NewShortSkillTierCategories(
	categories []blizzardv2.SkillTierCategory,
	providedLocale locale.Locale,
	recipeIconUrls map[blizzardv2.RecipeId]string,
) []ShortSkillTierCategory {
	out := make([]ShortSkillTierCategory, len(categories))
	for i, category := range categories {
		out[i] = ShortSkillTierCategory{
			Name:    category.Name.FindOr(providedLocale, ""),
			Recipes: NewShortSkillTierCategoryRecipes(category.Recipes, providedLocale, recipeIconUrls),
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
	recipeIconUrls map[blizzardv2.RecipeId]string,
) ShortSkillTier {
	return ShortSkillTier{
		Id:   skillTier.BlizzardMeta.Id,
		Name: skillTier.BlizzardMeta.Name.FindOr(providedLocale, ""),
		Categories: NewShortSkillTierCategories(
			skillTier.BlizzardMeta.Categories,
			providedLocale,
			recipeIconUrls,
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
