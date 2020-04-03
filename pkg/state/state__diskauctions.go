package state

import (
	"time"

	"github.com/sirupsen/logrus"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/diskstore"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type DiskAuctionsState struct {
	BlizzardState BlizzardState
	RegionsState  RegionsState

	DiskStore diskstore.DiskStore
}

type CollectAuctionsResult struct {
	Tuple        blizzardv2.RegionConnectedRealmTuple
	ItemIds      blizzardv2.ItemIds
	LastModified time.Time
}

type CollectAuctionsResults struct {
	ItemIds          blizzardv2.ItemIds
	RegionTimestamps sotah.RegionTimestamps
}

func (sta *DiskAuctionsState) CollectAuctions() (blizzardv2.ItemIds, error) {
	// spinning up workers
	aucsOutJobs := sta.BlizzardState.ResolveAuctions(sta.RegionsState.RegionComposites.ToDownloadTuples())
	storeAucsInJobs := make(chan diskstore.WriteAuctionsWithTuplesInJob)
	storeAucsOutJobs := sta.DiskStore.WriteAuctionsWithTuples(storeAucsInJobs)
	resultsInJob := make(chan CollectAuctionsResult)
	resultsOutJob := make(chan CollectAuctionsResults)

	// interpolating resolve-auctions-out jobs into store-auctions-in jobs
	go func() {
		for aucsOutJob := range aucsOutJobs {
			if aucsOutJob.Err != nil {
				logging.WithFields(aucsOutJob.ToLogrusFields()).Error("failed to fetch auctions")

				continue
			}

			if !aucsOutJob.IsNew {
				logging.WithFields(logrus.Fields{
					"region":             aucsOutJob.Tuple.RegionName,
					"connected-realm-id": aucsOutJob.Tuple.ConnectedRealmId,
				}).Info("auctions fetched successfully but no new results were found")

				continue
			}

			storeAucsInJobs <- diskstore.WriteAuctionsWithTuplesInJob{
				Tuple:    aucsOutJob.Tuple.RegionConnectedRealmTuple,
				Auctions: aucsOutJob.AuctionsResponse.Auctions,
			}
			resultsInJob <- CollectAuctionsResult{
				Tuple:        aucsOutJob.Tuple.RegionConnectedRealmTuple,
				ItemIds:      aucsOutJob.AuctionsResponse.Auctions.ItemIds(),
				LastModified: aucsOutJob.LastModified,
			}
		}

		close(storeAucsInJobs)
		close(resultsInJob)
	}()

	// spinning up a worker for receiving results from auctions-out worker
	go func() {
		results := CollectAuctionsResults{
			ItemIds:          blizzardv2.ItemIds{},
			RegionTimestamps: sotah.RegionTimestamps{},
		}
		for job := range resultsInJob {
			// loading last-modified in
			results.RegionTimestamps = results.RegionTimestamps.SetDownloaded(
				job.Tuple.RegionName,
				job.Tuple.ConnectedRealmId,
				job.LastModified,
			)

			// loading item-ids in
			results.ItemIds = results.ItemIds.Merge(job.ItemIds)
		}

		resultsOutJob <- results
		close(resultsOutJob)
	}()

	// waiting for store-auctions results to drain out
	for storeAucsOutJob := range storeAucsOutJobs {
		if storeAucsOutJob.Err != nil {
			logging.WithFields(storeAucsOutJob.ToLogrusFields()).Error("failed to store auctions")

			return blizzardv2.ItemIds{}, storeAucsOutJob.Err
		}
	}

	// waiting for item-ids to drain out
	results := <-resultsOutJob

	// optionally updating local state
	if !results.RegionTimestamps.IsZero() {
		sta.RegionsState.RegionComposites = sta.RegionsState.RegionComposites.Receive(results.RegionTimestamps)
	}

	return results.ItemIds, nil
}
