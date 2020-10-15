package disk

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func (c Client) Collect() error {
	startTime := time.Now()
	logging.Info("calling DiskCollector.Collect()")

	results, err := c.collectAuctions()
	if err != nil {
		return err
	}

	if err := c.CallLiveAuctionsIntake(results.tuples.RegionConnectedRealmTuples()); err != nil {
		return err
	}

	if err := c.CallPricelistHistoryIntake(results.tuples); err != nil {
		return err
	}

	if err := c.CallItemsIntake(results.itemIds); err != nil {
		return err
	}

	if err := c.CallPetsIntake(); err != nil {
		return err
	}

	logging.WithField(
		"duration-in-ms",
		time.Since(startTime).Milliseconds(),
	).Info("finished calling DiskCollector.Collect()")

	return nil
}
