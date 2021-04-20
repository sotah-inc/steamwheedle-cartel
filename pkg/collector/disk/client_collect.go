package disk

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func (c Client) Collect() error {
	startTime := time.Now()
	logging.Info("calling DiskCollector.Collect()")

	//collectAuctionsResults, err := c.collectAuctions()
	//if err != nil {
	//	return err
	//}
	//
	//if err := c.CallLiveAuctionsIntake(
	//	collectAuctionsResults.tuples.RegionConnectedRealmTuples(),
	//); err != nil {
	//	return err
	//}
	//
	//if err := c.CallItemPricesIntake(collectAuctionsResults.tuples); err != nil {
	//	return err
	//}
	//
	//if err := c.CallPetsIntake(); err != nil {
	//	return err
	//}
	//
	if err := c.CallProfessionsIntake(); err != nil {
		return err
	}

	if err := c.CallSkillTiersIntake(); err != nil {
		return err
	}

	recipesItemIds, err := c.CallRecipesIntake()
	if err != nil {
		return err
	}

	logging.WithField("recipe-item-ids", recipesItemIds).Info("receive recipe-items")

	//if err := c.CallRecipePricesIntake(collectAuctionsResults.tuples); err != nil {
	//	return err
	//}
	//
	//if err := c.CallStatsIntake(collectAuctionsResults.tuples); err != nil {
	//	return err
	//}
	//
	//if err := c.CallPrunePricelistHistories(); err != nil {
	//	return err
	//}
	//
	//// resolving next item-ids from auctions and recipes intake
	//nextItemIds := blizzardv2.ItemIdsMap{}
	//for _, id := range collectAuctionsResults.itemIds {
	//	nextItemIds[id] = struct{}{}
	//}
	//for _, id := range recipesItemIds {
	//	nextItemIds[id] = struct{}{}
	//}
	//
	//if err := c.CallItemsIntake(nextItemIds.ToItemIds()); err != nil {
	//	return err
	//}
	//
	//if err := c.CallTokenHistoryIntake(); err != nil {
	//	return err
	//}

	if err := c.CallEnchantingRecipeCorrelation(); err != nil {
		return err
	}

	logging.WithField(
		"duration-in-ms",
		time.Since(startTime).Milliseconds(),
	).Info("finished calling DiskCollector.Collect()")

	return nil
}
