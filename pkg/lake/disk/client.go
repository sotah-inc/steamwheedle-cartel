package disk

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type NewClientOptions struct {
	CacheDir     string
	RegionNames  []blizzardv2.RegionName
	GameVersions []gameversion.GameVersion
	ResolveItems func(
		version gameversion.GameVersion,
		ids blizzardv2.ItemIds,
	) chan blizzardv2.GetItemsOutJob
	ResolveItemMedias  func(in chan blizzardv2.GetItemMediasInJob) chan blizzardv2.GetItemMediasOutJob
	ResolvePets        func(blacklist []blizzardv2.PetId) (chan blizzardv2.GetAllPetsJob, error)
	ResolveProfessions func(
		blacklist []blizzardv2.ProfessionId,
	) (chan blizzardv2.GetAllProfessionsJob, error)
	ResolveProfessionMedias func(
		in chan blizzardv2.GetProfessionMediasInJob,
	) chan blizzardv2.GetProfessionMediasOutJob
	ResolveSkillTiers func(
		professionId blizzardv2.ProfessionId,
		idList []blizzardv2.SkillTierId,
	) chan blizzardv2.GetAllSkillTiersJob
	ResolveRecipes      func(group blizzardv2.RecipesGroup) chan blizzardv2.GetRecipesOutJob
	ResolveRecipeMedias func(
		in chan blizzardv2.GetRecipeMediasInJob,
	) chan blizzardv2.GetRecipeMediasOutJob
	PrimarySkillTiers  map[string][]blizzardv2.SkillTierId
	ResolveItemClasses func() ([]blizzardv2.ItemClassResponse, error)
}

func NewClient(opts NewClientOptions) (Client, error) {
	dirList := []string{opts.CacheDir, fmt.Sprintf("%s/auctions", opts.CacheDir)}
	for _, name := range opts.RegionNames {
		dirList = append(dirList, fmt.Sprintf("%s/auctions/%s", opts.CacheDir, name))

		for _, version := range opts.GameVersions {
			dirList = append(
				dirList,
				fmt.Sprintf("%s/auctions/%s/%s", opts.CacheDir, name, version),
			)
		}
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
		primarySkillTiers:       opts.PrimarySkillTiers,
		resolveItemClasses:      opts.ResolveItemClasses,
	}, nil
}

type Client struct {
	cacheDir     string
	resolveItems func(
		version gameversion.GameVersion,
		ids blizzardv2.ItemIds,
	) chan blizzardv2.GetItemsOutJob
	resolveItemMedias  func(in chan blizzardv2.GetItemMediasInJob) chan blizzardv2.GetItemMediasOutJob
	resolvePets        func(blacklist []blizzardv2.PetId) (chan blizzardv2.GetAllPetsJob, error)
	resolveProfessions func(
		blacklist []blizzardv2.ProfessionId,
	) (chan blizzardv2.GetAllProfessionsJob, error)
	resolveProfessionMedias func(
		in chan blizzardv2.GetProfessionMediasInJob,
	) chan blizzardv2.GetProfessionMediasOutJob
	resolveSkillTiers func(
		professionId blizzardv2.ProfessionId,
		idList []blizzardv2.SkillTierId,
	) chan blizzardv2.GetAllSkillTiersJob
	resolveRecipes      func(group blizzardv2.RecipesGroup) chan blizzardv2.GetRecipesOutJob
	resolveRecipeMedias func(
		in chan blizzardv2.GetRecipeMediasInJob,
	) chan blizzardv2.GetRecipeMediasOutJob
	primarySkillTiers  map[string][]blizzardv2.SkillTierId
	resolveItemClasses func() ([]blizzardv2.ItemClassResponse, error)
}
