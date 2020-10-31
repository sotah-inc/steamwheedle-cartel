package blizzardv2

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

type ProfessionSkillTierId int

type ProfessionSkillTierCategoryRecipe struct {
	Key  HrefReference  `json:"key"`
	Name locale.Mapping `json:"name"`
	Id   RecipeId       `json:"recipe"`
}

type ProfessionSkillTierCategory struct {
	Name    locale.Mapping                      `json:"name"`
	Recipes []ProfessionSkillTierCategoryRecipe `json:"recipes"`
}

type ProfessionSkillTier struct {
	LinksBase
	Id                ProfessionSkillTierId         `json:"id"`
	Name              locale.Mapping                `json:"name"`
	MinimumSkillLevel int                           `json:"minimum_skill_level"`
	MaximumSkillLevel int                           `json:"maximum_skill_level"`
	Categories        []ProfessionSkillTierCategory `json:"categories"`
}
