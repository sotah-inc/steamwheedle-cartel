package blizzardv2

type RecipesGroup map[ProfessionId]SkillTiersGroup

func (group RecipesGroup) FilterOut(ids RecipeIds) RecipesGroup {
	out := RecipesGroup{}
	for professionId, skillTierGroup := range group {
		resultSkillTierGroup := skillTierGroup.FilterOut(ids)
		if resultSkillTierGroup.IsZero() {
			continue
		}

		out[professionId] = resultSkillTierGroup
	}

	return out
}

func (group RecipesGroup) IsZero() bool {
	for _, skillTierGroup := range group {
		if !skillTierGroup.IsZero() {
			return false
		}
	}

	return true
}

func (group RecipesGroup) TotalRecipes() int {
	out := 0
	for _, skillTierGroup := range group {
		for _, recipeIds := range skillTierGroup {
			out += len(recipeIds)
		}
	}

	return out
}

type SkillTiersGroup map[SkillTierId]RecipeIds

func (group SkillTiersGroup) IsZero() bool {
	for _, recipeIds := range group {
		if !recipeIds.IsZero() {
			return false
		}
	}

	return true
}

func (group SkillTiersGroup) FilterOut(ids RecipeIds) SkillTiersGroup {
	out := SkillTiersGroup{}
	for skillTierId, recipeIds := range group {
		resultRecipeIds := recipeIds.Subtract(ids)
		if resultRecipeIds.IsZero() {
			continue
		}

		out[skillTierId] = resultRecipeIds
	}

	return out
}
