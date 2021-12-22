package gcloud

import (
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func NewClient() (Client, error) {
	return Client{}, errors.New("NYI")
}

type Client struct{}

func (c Client) GetEncodedAuctionsByTuples(
	tuples blizzardv2.RegionVersionConnectedRealmTuples,
) chan BaseLake.GetEncodedAuctionsByTuplesJob {
	//TODO implement me
	panic("implement me")
}

func (c Client) GetEncodedStatsByTuples(
	tuples blizzardv2.LoadConnectedRealmTuples,
) chan BaseLake.GetEncodedStatsByTuplesJob {
	//TODO implement me
	panic("implement me")
}

func (c Client) GetEncodedItemPricesByTuples(
	tuples blizzardv2.LoadConnectedRealmTuples,
) chan BaseLake.GetEncodedItemPricesByTuplesJob {
	//TODO implement me
	panic("implement me")
}

func (c Client) GetEncodedRecipePricesByTuples(
	mRecipes sotah.MiniRecipes,
	tuples blizzardv2.LoadConnectedRealmTuples,
) chan BaseLake.GetEncodedRecipePricesByTuplesJob {
	//TODO implement me
	panic("implement me")
}

func (c Client) WriteAuctionsWithTuples(
	in chan BaseLake.WriteAuctionsWithTuplesInJob,
) chan BaseLake.WriteAuctionsWithTuplesOutJob {
	//TODO implement me
	panic("implement me")
}

func (c Client) NewWriteAuctionsWithTuplesInJob(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
	auctions sotah.MiniAuctionList,
) BaseLake.WriteAuctionsWithTuplesInJob {
	//TODO implement me
	panic("implement me")
}

func (c Client) GetEncodedItems(
	version gameversion.GameVersion,
	ids blizzardv2.ItemIds,
) (chan BaseLake.GetEncodedItemJob, chan []blizzardv2.ItemId) {
	//TODO implement me
	panic("implement me")
}

func (c Client) GetEncodedPets(
	blacklist []blizzardv2.PetId,
) (chan BaseLake.GetEncodedPetJob, error) {
	//TODO implement me
	panic("implement me")
}

func (c Client) GetEncodedProfessions(
	blacklist []blizzardv2.ProfessionId,
) (chan BaseLake.GetEncodedProfessionJob, error) {
	//TODO implement me
	panic("implement me")
}

func (c Client) GetEncodedSkillTiers(
	professionId blizzardv2.ProfessionId,
	idList []blizzardv2.SkillTierId,
) chan BaseLake.GetEncodedSkillTierJob {
	//TODO implement me
	panic("implement me")
}

func (c Client) GetEncodedRecipes(
	group blizzardv2.RecipesGroup,
) chan BaseLake.GetEncodedRecipeJob {
	//TODO implement me
	panic("implement me")
}

func (c Client) GetEncodedRegionStats(
	tuple blizzardv2.RegionVersionTuple,
	ids []blizzardv2.ConnectedRealmId,
) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (c Client) GetEncodedItemClasses() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
