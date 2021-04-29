package itemrecipekind

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

type ItemRecipeKind string

const (
	CraftedBy  ItemRecipeKind = "crafted-by"
	ReagentFor ItemRecipeKind = "reagent-for"
	Teaches    ItemRecipeKind = "teaches"
)

func IsValid(providedKind ItemRecipeKind) bool {
	kinds := []ItemRecipeKind{
		CraftedBy,
		Teaches,
		ReagentFor,
	}
	for _, kind := range kinds {
		if kind == providedKind {
			return true
		}
	}

	return false
}

func NewKindRecipesMap(kinds []ItemRecipeKind) KindRecipesMap {
	out := KindRecipesMap{}
	for _, kind := range kinds {
		out[kind] = blizzardv2.ItemRecipesMap{}
	}

	return out
}

type KindRecipesMap map[ItemRecipeKind]blizzardv2.ItemRecipesMap

func (krMap KindRecipesMap) Find(kind ItemRecipeKind) blizzardv2.ItemRecipesMap {
	found, ok := krMap[kind]
	if !ok {
		return blizzardv2.ItemRecipesMap{}
	}

	return found
}

func (krMap KindRecipesMap) Insert(
	kind ItemRecipeKind,
	irMap blizzardv2.ItemRecipesMap,
) KindRecipesMap {
	krMap[kind] = krMap.Find(kind).Merge(irMap)

	return krMap
}
