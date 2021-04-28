package itemrecipekind

type ItemRecipeKind string

const (
	CraftedBy  ItemRecipeKind = "crafted-by"
	ReagentFor ItemRecipeKind = "reagent-for"
	Teaches    ItemRecipeKind = "teaches"
)
