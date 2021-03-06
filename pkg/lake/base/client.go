package base

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type Client interface {
	GetEncodedAuctionsByTuples(
		tuples blizzardv2.RegionConnectedRealmTuples,
	) chan GetEncodedAuctionsByTuplesJob
	GetEncodedStatsByTuples(tuples blizzardv2.LoadConnectedRealmTuples) chan GetEncodedStatsByTuplesJob
	GetEncodedItemPricesByTuples(
		tuples blizzardv2.LoadConnectedRealmTuples,
	) chan GetEncodedItemPricesByTuplesJob
	GetEncodedRecipePricesByTuples(
		mRecipes sotah.MiniRecipes,
		tuples blizzardv2.LoadConnectedRealmTuples,
	) chan GetEncodedRecipePricesByTuplesJob
	WriteAuctionsWithTuples(in chan WriteAuctionsWithTuplesInJob) chan WriteAuctionsWithTuplesOutJob
	NewWriteAuctionsWithTuplesInJob(
		tuple blizzardv2.RegionConnectedRealmTuple,
		auctions sotah.MiniAuctionList,
	) WriteAuctionsWithTuplesInJob
	GetEncodedItems(ids blizzardv2.ItemIds) (chan GetEncodedItemJob, chan []blizzardv2.ItemId)
	GetEncodedPets(blacklist []blizzardv2.PetId) (chan GetEncodedPetJob, error)
	GetEncodedProfessions(blacklist []blizzardv2.ProfessionId) (chan GetEncodedProfessionJob, error)
	GetEncodedSkillTiers(
		professionId blizzardv2.ProfessionId,
		idList []blizzardv2.SkillTierId,
	) chan GetEncodedSkillTierJob
	GetEncodedRecipes(group blizzardv2.RecipesGroup) chan GetEncodedRecipeJob
	GetEncodedRegionStats(
		name blizzardv2.RegionName,
		ids []blizzardv2.ConnectedRealmId,
	) ([]byte, error)
}
