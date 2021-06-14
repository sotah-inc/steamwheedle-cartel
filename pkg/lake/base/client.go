package base

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type Client interface {
	GetEncodedAuctionsByTuples(
		tuples blizzardv2.RegionVersionConnectedRealmTuples,
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
		tuple blizzardv2.RegionVersionConnectedRealmTuple,
		auctions sotah.MiniAuctionList,
	) WriteAuctionsWithTuplesInJob
	GetEncodedItems(
		version gameversion.GameVersion,
		ids blizzardv2.ItemIds,
	) (chan GetEncodedItemJob, chan []blizzardv2.ItemId)
	GetEncodedPets(blacklist []blizzardv2.PetId) (chan GetEncodedPetJob, error)
	GetEncodedProfessions(blacklist []blizzardv2.ProfessionId) (chan GetEncodedProfessionJob, error)
	GetEncodedSkillTiers(
		professionId blizzardv2.ProfessionId,
		idList []blizzardv2.SkillTierId,
	) chan GetEncodedSkillTierJob
	GetEncodedRecipes(group blizzardv2.RecipesGroup) chan GetEncodedRecipeJob
	GetEncodedRegionStats(
		tuple blizzardv2.RegionVersionTuple,
		ids []blizzardv2.ConnectedRealmId,
	) ([]byte, error)
	GetEncodedItemClasses() ([]byte, error)
}
