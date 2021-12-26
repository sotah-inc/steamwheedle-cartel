package disk

import (
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseCollector "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/collector/base" // nolint:lll
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/statuskinds"
)

type collectAuctionsResult struct {
	tuple   blizzardv2.LoadConnectedRealmTuple
	itemIds blizzardv2.ItemIds
}

func (c Client) CollectAuctions() (BaseCollector.CollectAuctionsResults, error) {
	startTime := time.Now()
	logging.Info("calling DiskCollector.collectAuctions()")

	// spinning up workers
	aucsOutJobs, err := c.resolveAuctions()
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to produce chan for resolving auctions")

		return BaseCollector.CollectAuctionsResults{}, err
	}
	storeAucsInJobs := make(chan BaseLake.WriteAuctionsWithTuplesInJob)
	storeAucsOutJobs := c.lakeClient.WriteAuctionsWithTuples(storeAucsInJobs)
	resultsInJob := make(chan collectAuctionsResult)
	resultsOutJob := make(chan BaseCollector.CollectAuctionsResults)

	// interpolating resolve-auctions-out jobs into store-auctions-in jobs
	go func() {
		for aucsOutJob := range aucsOutJobs {
			if aucsOutJob.Err != nil {
				logging.WithFields(aucsOutJob.ToLogrusFields()).Error("failed to fetch auctions")

				continue
			}

			if !aucsOutJob.IsNew() {
				logging.WithFields(logrus.Fields{
					"region":             aucsOutJob.Tuple.RegionName,
					"connected-realm-id": aucsOutJob.Tuple.ConnectedRealmId,
				}).Info("auctions fetched successfully but no new results were found")

				continue
			}

			storeAucsInJobs <- c.lakeClient.NewWriteAuctionsWithTuplesInJob(
				aucsOutJob.Tuple.RegionVersionConnectedRealmTuple,
				sotah.NewMiniAuctionList(aucsOutJob.AuctionsResponse.Auctions),
			)
			resultsInJob <- collectAuctionsResult{
				tuple:   aucsOutJob.Tuple,
				itemIds: aucsOutJob.AuctionsResponse.Auctions.ItemIds(),
			}
		}

		close(storeAucsInJobs)
		close(resultsInJob)
	}()

	// spinning up a worker for receiving results from auctions-out worker
	go func() {
		results := BaseCollector.CollectAuctionsResults{
			VersionItems:            blizzardv2.VersionItemsMap{},
			RegionVersionTimestamps: sotah.RegionVersionTimestamps{},
			Tuples:                  blizzardv2.LoadConnectedRealmTuples{},
		}
		for job := range resultsInJob {
			// loading last-modified in
			results.RegionVersionTimestamps = results.RegionVersionTimestamps.SetTimestamp(
				job.tuple.RegionVersionConnectedRealmTuple,
				statuskinds.Downloaded,
				job.tuple.LastModified,
			)

			// loading item-ids in
			results.VersionItems = results.VersionItems.Insert(job.tuple.Version, job.itemIds)

			// loading tuple in
			results.Tuples = append(results.Tuples, job.tuple)
		}

		resultsOutJob <- results
		close(resultsOutJob)
	}()

	// waiting for store-auctions results to drain out
	totalPersisted := 0
	for storeAucsOutJob := range storeAucsOutJobs {
		if storeAucsOutJob.Err() != nil {
			logging.WithFields(storeAucsOutJob.ToLogrusFields()).Error("failed to store auctions")

			return BaseCollector.CollectAuctionsResults{}, storeAucsOutJob.Err()
		}

		totalPersisted += 1
	}

	// waiting for item-ids to drain out
	results := <-resultsOutJob

	// optionally updating region state

	logging.WithField(
		"timestamps",
		results.RegionVersionTimestamps,
	).Info("CollectAuctions() received region-version timestamps")

	if !results.RegionVersionTimestamps.IsZero() {
		if err := c.receiveRegionTimestamps(results.RegionVersionTimestamps); err != nil {
			logging.WithField("error", err.Error()).Error("failed to receive timestamps")

			return BaseCollector.CollectAuctionsResults{}, err
		}
	}

	logging.WithFields(logrus.Fields{
		"total":          totalPersisted,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in collect-auctions")

	return results, nil
}
