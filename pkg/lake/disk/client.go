package disk

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type NewClientOptions struct {
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

func NewClient(opts NewClientOptions) (Client, error) {
	dirList := []string{opts.CacheDir, fmt.Sprintf("%s/auctions", opts.CacheDir)}
	for _, name := range opts.RegionNames {
		dirList = append(dirList, fmt.Sprintf("%s/auctions/%s", opts.CacheDir, name))
	}

	// ensuring related dirs exist
	if err := util.EnsureDirsExist(dirList); err != nil {
		return Client{}, err
	}

	return Client{
		cacheDir:                opts.CacheDir,
		resolveItems:            opts.ResolveItems,
		resolveItemMedias:       opts.ResolveItemMedias,
		resolvePets:             opts.ResolvePets,
		resolveProfessions:      opts.ResolveProfessions,
		resolveProfessionMedias: opts.ResolveProfessionMedias,
		resolveSkillTiers:       opts.ResolveSkillTiers,
		resolveRecipes:          opts.ResolveRecipes,
		resolveRecipeMedias:     opts.ResolveRecipeMedias,
		primaryList:             opts.PrimaryList,
	}, nil
}

type Client struct {
	cacheDir                string
	resolveItems            func(ids blizzardv2.ItemIds) chan blizzardv2.GetItemsOutJob
	resolveItemMedias       func(in chan blizzardv2.GetItemMediasInJob) chan blizzardv2.GetItemMediasOutJob
	resolvePets             func(blacklist []blizzardv2.PetId) (chan blizzardv2.GetAllPetsJob, error)
	resolveProfessions      func(blacklist []blizzardv2.ProfessionId) (chan blizzardv2.GetAllProfessionsJob, error)
	resolveProfessionMedias func(in chan blizzardv2.GetProfessionMediasInJob) chan blizzardv2.GetProfessionMediasOutJob
	resolveSkillTiers       func(
		professionId blizzardv2.ProfessionId,
		idList []blizzardv2.SkillTierId,
	) chan blizzardv2.GetAllSkillTiersJob
	resolveRecipes      func(ids []blizzardv2.RecipeId) chan blizzardv2.GetRecipesJob
	resolveRecipeMedias func(in chan blizzardv2.GetRecipeMediasInJob) chan blizzardv2.GetRecipeMediasOutJob
	primaryList         []blizzardv2.ProfessionId
}
