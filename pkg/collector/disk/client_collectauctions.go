package disk

import (
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseCollector "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/collector/base"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type CollectAuctionsResult struct {
	tuple   blizzardv2.LoadConnectedRealmTuple
	itemIds blizzardv2.ItemIds
}

func (c CollectAuctionsResult) Tuple() blizzardv2.LoadConnectedRealmTuple { return c.tuple }
func (c CollectAuctionsResult) ItemIds() blizzardv2.ItemIds               { return c.itemIds }

type CollectAuctionsResults struct {
	itemIds          blizzardv2.ItemIds
	regionTimestamps sotah.RegionTimestamps
	tuples           blizzardv2.LoadConnectedRealmTuples
}

func (c CollectAuctionsResults) ItemIds() blizzardv2.ItemIds                 { return c.itemIds }
func (c CollectAuctionsResults) RegionTimestamps() sotah.RegionTimestamps    { return c.regionTimestamps }
func (c CollectAuctionsResults) Tuples() blizzardv2.LoadConnectedRealmTuples { return c.tuples }

func (c Client) CollectAuctions() (BaseCollector.CollectAuctionsResults, error) {
	startTime := time.Now()

	// spinning up workers
	aucsOutJobs := c.resolveAuctions()
	storeAucsInJobs := make(chan BaseLake.WriteAuctionsWithTuplesInJob)
	storeAucsOutJobs := c.lakeClient.WriteAuctionsWithTuples(storeAucsInJobs)
	resultsInJob := make(chan BaseCollector.CollectAuctionsResult)
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
				aucsOutJob.Tuple.RegionConnectedRealmTuple,
				sotah.NewMiniAuctionList(aucsOutJob.AuctionsResponse.Auctions),
			)
			resultsInJob <- CollectAuctionsResult{
				tuple:   aucsOutJob.Tuple,
				itemIds: aucsOutJob.AuctionsResponse.Auctions.ItemIds(),
			}
		}

		close(storeAucsInJobs)
		close(resultsInJob)
	}()

	// spinning up a worker for receiving results from auctions-out worker
	go func() {
		results := CollectAuctionsResults{
			itemIds:          blizzardv2.ItemIds{},
			regionTimestamps: sotah.RegionTimestamps{},
			tuples:           blizzardv2.LoadConnectedRealmTuples{},
		}
		for job := range resultsInJob {
			// loading last-modified in
			results.regionTimestamps = results.RegionTimestamps().SetDownloaded(
				job.Tuple().RegionConnectedRealmTuple,
				job.Tuple().LastModified,
			)

			// loading item-ids in
			results.itemIds = results.itemIds.Merge(job.ItemIds())

			// loading tuple in
			results.tuples = append(results.Tuples(), job.Tuple())
		}

		results.itemIds = func() blizzardv2.ItemIds {
			if len(results.itemIds) < 5 {
				return results.itemIds
			}

			return results.itemIds[:5]
		}()

		resultsOutJob <- results
		close(resultsOutJob)
	}()

	// waiting for store-auctions results to drain out
	totalPersisted := 0
	for storeAucsOutJob := range storeAucsOutJobs {
		if storeAucsOutJob.Err() != nil {
			logging.WithFields(storeAucsOutJob.ToLogrusFields()).Error("failed to store auctions")

			return CollectAuctionsResults{}, storeAucsOutJob.Err()
		}

		totalPersisted += 1
	}

	// waiting for item-ids to drain out
	results := <-resultsOutJob

	// optionally updating region state
	if !results.RegionTimestamps().IsZero() {
		c.receiveRegionTimestamps(results.RegionTimestamps())
	}

	logging.WithFields(logrus.Fields{
		"total":          totalPersisted,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in collect-auctions")

	return results, nil
}
