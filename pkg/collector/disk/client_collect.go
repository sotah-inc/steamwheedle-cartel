package disk

import (
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
)

func (c Client) Collect() error {
	startTime := time.Now()
	logging.Info("calling DiskCollector.Collect()")

	collectAuctionsResults, err := c.CollectAuctions()
	if err != nil {
		return err
	}

	if err := c.CallLiveAuctionsIntake(state.IntakeRequest{
		Tuples: collectAuctionsResults.Tuples.RegionVersionConnectedRealmTuples(),
	}); err != nil {
		return err
	}

	if err := c.CallItemPricesIntake(collectAuctionsResults.Tuples); err != nil {
		return err
	}

	if err := c.CallPetsIntake(); err != nil {
		return err
	}

	if err := c.CallProfessionsIntake(); err != nil {
		return err
	}

	if err := c.CallSkillTiersIntake(); err != nil {
		return err
	}

	recipesIntakeResponse, err := c.CallRecipesIntake()
	if err != nil {
		return err
	}

	if err := c.CallRecipePricesIntake(collectAuctionsResults.Tuples); err != nil {
		return err
	}

	if err := c.CallStatsIntake(collectAuctionsResults.Tuples); err != nil {
		return err
	}

	if err := c.CallPrunePricelistHistories(); err != nil {
		return err
	}

	// resolving next item-ids from auctions and recipes intake
	logging.WithFields(logrus.Fields{
		"recipe-item-ids": len(recipesIntakeResponse.RecipeItemIds),
	}).Info("DID NOT COMBINE RECIPE ITEM-IDS IN INTAKE RESPONSE")
	itemIntakeResponse, err := c.CallItemsIntake(collectAuctionsResults.VersionItems)
	if err != nil {
		return err
	}

	if err := c.CallTokenHistoryIntake(); err != nil {
		return err
	}

	if recipesIntakeResponse.TotalPersisted > 0 || itemIntakeResponse.TotalPersisted > 0 {
		recipeItemStartTime := time.Now()

		logging.Info("starting item-recipes correlation tasks")

		if err := c.CallEnchantingRecipeCorrelation(); err != nil {
			return err
		}

		if err := c.CallRecipeItemCorrelation(); err != nil {
			return err
		}

		logging.WithField(
			"duration-in-ms",
			time.Since(recipeItemStartTime).Milliseconds(),
		).Info("finished calling item-recipes correlation tasks")
	}

	logging.WithField(
		"duration-in-ms",
		time.Since(startTime).Milliseconds(),
	).Info("finished calling DiskCollector.Collect()")

	return nil
}
