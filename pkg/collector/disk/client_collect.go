package disk

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func (c Client) Collect() error {
	startTime := time.Now()
	logging.Info("calling DiskCollector.Collect()")

	collectAuctionsResults, err := c.collectAuctions()
	if err != nil {
		return err
	}

	if err := c.CallLiveAuctionsIntake(collectAuctionsResults.tuples.RegionConnectedRealmTuples()); err != nil {
		return err
	}

	if err := c.CallPricelistHistoryIntake(collectAuctionsResults.tuples); err != nil {
		return err
	}

	if err := c.CallItemsIntake(collectAuctionsResults.itemIds); err != nil {
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

	if err := c.CallRecipesIntake(); err != nil {
		return err
	}

	logging.WithField(
		"duration-in-ms",
		time.Since(startTime).Milliseconds(),
	).Info("finished calling DiskCollector.Collect()")

	return nil
}
