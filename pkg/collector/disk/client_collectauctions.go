package disk

import (
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type collectAuctionsResult struct {
	tuple   blizzardv2.LoadConnectedRealmTuple
	itemIds blizzardv2.ItemIds
}

type collectAuctionsResults struct {
	itemIds          blizzardv2.ItemIds
	regionTimestamps sotah.RegionTimestamps
	tuples           blizzardv2.LoadConnectedRealmTuples
}

const itemIdsLimit = 250

func (c Client) collectAuctions() (collectAuctionsResults, error) {
	startTime := time.Now()
	logging.Info("calling DiskCollector.collectAuctions()")

	// spinning up workers
	aucsOutJobs := c.resolveAuctions()
	storeAucsInJobs := make(chan BaseLake.WriteAuctionsWithTuplesInJob)
	storeAucsOutJobs := c.lakeClient.WriteAuctionsWithTuples(storeAucsInJobs)
	resultsInJob := make(chan collectAuctionsResult)
	resultsOutJob := make(chan collectAuctionsResults)

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
				aucsOutJob.Tuple.RegionConnectedRealmTuple,
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
		results := collectAuctionsResults{
			itemIds:          blizzardv2.ItemIds{},
			regionTimestamps: sotah.RegionTimestamps{},
			tuples:           blizzardv2.LoadConnectedRealmTuples{},
		}
		for job := range resultsInJob {
			// loading last-modified in
			results.regionTimestamps = results.regionTimestamps.SetDownloaded(
				job.tuple.RegionConnectedRealmTuple,
				job.tuple.LastModified,
			)

			// loading item-ids in
			results.itemIds = results.itemIds.Merge(job.itemIds)

			// loading tuple in
			results.tuples = append(results.tuples, job.tuple)
		}

		results.itemIds = func() blizzardv2.ItemIds {
			if len(results.itemIds) < itemIdsLimit {
				return results.itemIds
			}

			return results.itemIds[:itemIdsLimit]
		}()

		resultsOutJob <- results
		close(resultsOutJob)
	}()

	// waiting for store-auctions results to drain out
	totalPersisted := 0
	for storeAucsOutJob := range storeAucsOutJobs {
		if storeAucsOutJob.Err() != nil {
			logging.WithFields(storeAucsOutJob.ToLogrusFields()).Error("failed to store auctions")

			return collectAuctionsResults{}, storeAucsOutJob.Err()
		}

		totalPersisted += 1
	}

	// waiting for item-ids to drain out
	results := <-resultsOutJob

	// optionally updating region state
	if !results.regionTimestamps.IsZero() {
		c.receiveRegionTimestamps(results.regionTimestamps)
	}

	logging.WithFields(logrus.Fields{
		"total":          totalPersisted,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in collect-auctions")

	return results, nil
}
