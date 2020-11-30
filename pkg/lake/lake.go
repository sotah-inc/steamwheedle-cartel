package lake

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	DiskLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/disk"
)

type NewClientOptions struct {
	UseGCloud               bool
	CacheDir                string
	RegionNames             []blizzardv2.RegionName
	ResolveItems            func(ids blizzardv2.ItemIds) chan blizzardv2.GetItemsOutJob
	ResolveItemMedias       func(in chan blizzardv2.GetItemMediasInJob) chan blizzardv2.GetItemMediasOutJob
	ResolvePets             func(blacklist []blizzardv2.PetId) (chan blizzardv2.GetAllPetsJob, error)
	ResolveProfessions      func(blacklist []blizzardv2.ProfessionId) (chan blizzardv2.GetAllProfessionsJob, error)
	ResolveProfessionMedias func(in chan blizzardv2.GetProfessionMediasInJob) chan blizzardv2.GetProfessionMediasOutJob
	ResolveSkillTiers       func(
		professionId blizzardv2.ProfessionId,
		idList []blizzardv2.SkillTierId,
	) chan blizzardv2.GetAllSkillTiersJob
	ResolveRecipes      func(ids []blizzardv2.RecipeId) chan blizzardv2.GetRecipesJob
	ResolveRecipeMedias func(in chan blizzardv2.GetRecipeMediasInJob) chan blizzardv2.GetRecipeMediasOutJob
	PrimaryList         []blizzardv2.ProfessionId
}

func NewClient(opts NewClientOptions) (BaseLake.Client, error) {
	if opts.UseGCloud {
		return DiskLake.NewClient(DiskLake.NewClientOptions{
			CacheDir:                opts.CacheDir,
			ResolveItems:            opts.ResolveItems,
			ResolveItemMedias:       opts.ResolveItemMedias,
			RegionNames:             opts.RegionNames,
			ResolvePets:             opts.ResolvePets,
			ResolveProfessions:      opts.ResolveProfessions,
			ResolveProfessionMedias: opts.ResolveProfessionMedias,
			ResolveSkillTiers:       opts.ResolveSkillTiers,
			ResolveRecipes:          opts.ResolveRecipes,
			ResolveRecipeMedias:     opts.ResolveRecipeMedias,
			PrimaryList:             opts.PrimaryList,
		})
	}

	return DiskLake.NewClient(DiskLake.NewClientOptions{
		CacheDir:                opts.CacheDir,
		ResolveItems:            opts.ResolveItems,
		ResolveItemMedias:       opts.ResolveItemMedias,
		RegionNames:             opts.RegionNames,
		ResolvePets:             opts.ResolvePets,
		ResolveProfessions:      opts.ResolveProfessions,
		ResolveProfessionMedias: opts.ResolveProfessionMedias,
		ResolveSkillTiers:       opts.ResolveSkillTiers,
		ResolveRecipes:          opts.ResolveRecipes,
		ResolveRecipeMedias:     opts.ResolveRecipeMedias,
		PrimaryList:             opts.PrimaryList,
	})
}
