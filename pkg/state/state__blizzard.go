package state

import (
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"

	"github.com/sirupsen/logrus"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func NewBlizzardState(config blizzardv2.ClientConfig) (BlizzardState, error) {
	client, err := blizzardv2.NewClient(config)
	if err != nil {
		return BlizzardState{}, err
	}

	if !client.IsValid() {
		logging.WithField("source", "NewBlizzardState").Error("client was not valid")

		return BlizzardState{}, errors.New("client was not valid")
	}

	return BlizzardState{BlizzardClient: client}, nil
}

type BlizzardState struct {
	BlizzardClient *blizzardv2.Client
}

func (sta BlizzardState) ResolveConnectedRealms(
	region sotah.Region,
	version gameversion.GameVersion,
	blacklist []blizzardv2.ConnectedRealmId,
	whitelist *blizzardv2.RealmSlugs,
) (chan blizzardv2.GetAllConnectedRealmsJob, error) {
	return blizzardv2.GetAllConnectedRealms(blizzardv2.GetAllConnectedRealmsOptions{
		GetConnectedRealmIndexURL: func() (string, error) {
			defaultUrl, err := blizzardv2.DefaultConnectedRealmIndexURL(
				region.Hostname,
				version,
			)
			if err != nil {
				return "", err
			}

			return sta.BlizzardClient.AppendAccessToken(defaultUrl)
		},
		GetConnectedRealmURL: sta.BlizzardClient.AppendAccessToken,
		Blacklist:            blacklist,
		RealmWhitelist:       whitelist,
	})
}

func (sta BlizzardState) ResolveItemClasses(
	regions sotah.RegionList,
) ([]blizzardv2.ItemClassResponse, error) {
	primaryRegion, err := regions.GetPrimaryRegion()
	if err != nil {
		return []blizzardv2.ItemClassResponse{}, err
	}

	return blizzardv2.GetAllItemClasses(blizzardv2.GetAllItemClassesOptions{
		GetItemClassIndexURL: func() (string, error) {
			return sta.BlizzardClient.AppendAccessToken(
				blizzardv2.DefaultGetItemClassIndexURL(primaryRegion.Hostname, primaryRegion.Name),
			)
		},
		GetItemClassURL: func(id itemclass.Id) (string, error) {
			return sta.BlizzardClient.AppendAccessToken(
				blizzardv2.DefaultGetItemClassURL(primaryRegion.Hostname, primaryRegion.Name, id),
			)
		},
	})
}

func (sta BlizzardState) ResolveTokens(
	regions sotah.RegionList,
) (map[blizzardv2.RegionName]blizzardv2.TokenResponse, error) {
	return blizzardv2.GetTokens(blizzardv2.GetTokensOptions{
		Tuples: func() []blizzardv2.RegionHostnameTuple {
			out := make([]blizzardv2.RegionHostnameTuple, len(regions))
			for i, region := range regions {
				out[i] = blizzardv2.RegionHostnameTuple{
					RegionName:     region.Name,
					RegionHostname: region.Hostname,
				}
			}

			return out
		}(),
		GetTokenInfoURL: func(regionHostname string, regionName blizzardv2.RegionName) (string, error) {
			return sta.BlizzardClient.AppendAccessToken(
				blizzardv2.DefaultGetTokenURL(regionHostname, regionName),
			)
		},
	})
}

func (sta BlizzardState) ResolveAuctions(
	tuples []blizzardv2.DownloadConnectedRealmTuple,
) chan blizzardv2.GetAuctionsJob {
	logging.WithField("tuples", len(tuples)).Info("resolving auctions with tuples")

	return blizzardv2.GetAuctions(blizzardv2.GetAuctionsOptions{
		Tuples: tuples,
		GetAuctionsURL: func(tuple blizzardv2.DownloadConnectedRealmTuple) (string, error) {
			return sta.BlizzardClient.AppendAccessToken(blizzardv2.DefaultGetAuctionsURL(tuple))
		},
	})
}

func (sta BlizzardState) ResolveItems(
	primaryRegion sotah.Region,
	version gameversion.GameVersion,
	ids blizzardv2.ItemIds,
) chan blizzardv2.GetItemsOutJob {
	logging.WithField("item-ids", len(ids)).Info("resolving item-ids")

	return blizzardv2.GetItems(blizzardv2.GetItemsOptions{
		GetItemURL: func(id blizzardv2.ItemId) (string, error) {
			dst, err := blizzardv2.DefaultGetItemURL(primaryRegion.Hostname, id, version)
			if err != nil {
				return "", err
			}

			return sta.BlizzardClient.AppendAccessToken(dst)
		},
		ItemIds: ids,
		Limit:   1000,
	})
}

func (sta BlizzardState) ResolveItemMedias(
	in chan blizzardv2.GetItemMediasInJob,
) chan blizzardv2.GetItemMediasOutJob {
	return blizzardv2.GetItemMedias(in, sta.BlizzardClient.AppendAccessToken)
}

func (sta BlizzardState) ResolvePets(
	primaryRegion sotah.Region,
	blacklist []blizzardv2.PetId,
) (chan blizzardv2.GetAllPetsJob, error) {
	logging.WithField("pet-ids", len(blacklist)).Info("resolving pets with blacklist")

	return blizzardv2.GetAllPets(blizzardv2.GetAllPetsOptions{
		GetPetIndexURL: func() (string, error) {
			return sta.BlizzardClient.AppendAccessToken(
				blizzardv2.DefaultPetIndexURL(primaryRegion.Hostname, primaryRegion.Name),
			)
		},
		GetPetURL: sta.BlizzardClient.AppendAccessToken,
		Blacklist: blacklist,
		Limit:     250,
	})
}

func (sta BlizzardState) ResolveProfessions(
	primaryRegion sotah.Region,
	blacklist []blizzardv2.ProfessionId,
) (chan blizzardv2.GetAllProfessionsJob, error) {
	logging.WithField("profession-ids", blacklist).Info("resolving professions with blacklist")

	return blizzardv2.GetAllProfessions(blizzardv2.GetAllProfessionsOptions{
		GetProfessionIndexURL: func() (string, error) {
			return sta.BlizzardClient.AppendAccessToken(
				blizzardv2.DefaultProfessionIndexURL(primaryRegion.Hostname, primaryRegion.Name),
			)
		},
		GetProfessionURL: sta.BlizzardClient.AppendAccessToken,
		Blacklist:        blacklist,
	})
}

func (sta BlizzardState) ResolveProfessionMedias(
	in chan blizzardv2.GetProfessionMediasInJob,
) chan blizzardv2.GetProfessionMediasOutJob {
	return blizzardv2.GetProfessionsMedias(in, sta.BlizzardClient.AppendAccessToken)
}

func (sta BlizzardState) ResolveSkillTiers(
	primaryRegion sotah.Region,
	professionId blizzardv2.ProfessionId,
	idList []blizzardv2.SkillTierId,
) chan blizzardv2.GetAllSkillTiersJob {
	logging.WithFields(logrus.Fields{
		"profession":   professionId,
		"provided-ids": len(idList),
	}).Info("resolving skill-tiers")

	return blizzardv2.GetAllSkillTiers(blizzardv2.GetAllSkillTiersOptions{
		GetSkillTierURL: func(id blizzardv2.SkillTierId) (string, error) {
			return sta.BlizzardClient.AppendAccessToken(blizzardv2.DefaultSkillTierURL(
				primaryRegion.Hostname,
				professionId,
				id,
				primaryRegion.Name,
			))
		},
		SkillTierIdList: idList,
	})
}

func (sta BlizzardState) ResolveRecipes(
	primaryRegion sotah.Region,
	recipesGroup blizzardv2.RecipesGroup,
) chan blizzardv2.GetRecipesOutJob {
	logging.WithField("recipe-ids", recipesGroup.TotalRecipes()).Info("resolving recipe-ids")

	return blizzardv2.GetRecipes(blizzardv2.GetRecipesOptions{
		GetRecipeURL: func(id blizzardv2.RecipeId) (string, error) {
			return sta.BlizzardClient.AppendAccessToken(
				blizzardv2.DefaultGetRecipeURL(primaryRegion.Hostname, id, primaryRegion.Name),
			)
		},
		RecipesGroup: recipesGroup,
		Limit:        250,
	})
}

func (sta BlizzardState) ResolveRecipeMedias(
	in chan blizzardv2.GetRecipeMediasInJob,
) chan blizzardv2.GetRecipeMediasOutJob {
	return blizzardv2.GetRecipeMedias(in, sta.BlizzardClient.AppendAccessToken)
}
