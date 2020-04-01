package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/diskstore"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type DiskAuctionsState struct {
	BlizzardState BlizzardState

	DiskStore diskstore.DiskStore
}

func (sta DiskAuctionsState) CollectAuctions(
	tuples []blizzardv2.RegionConnectedRealmTuple,
) (blizzardv2.ItemIds, error) {
	// spinning up workers
	aucsOutJobs := sta.BlizzardState.ResolveAuctions(tuples)
	storeAucsInJobs := make(chan diskstore.WriteAuctionsWithTuplesInJob)
	storeAucsOutJobs := sta.DiskStore.WriteAuctionsWithTuples(storeAucsInJobs)
	itemIdsInJobs := make(chan blizzardv2.ItemIds)
	itemIdsOutJob := make(chan blizzardv2.ItemIds)

	// interpolating resolve-auctions-out jobs into store-auctions-in jobs
	go func() {
		for aucsOutJob := range aucsOutJobs {
			if aucsOutJob.Err != nil {
				logging.WithFields(aucsOutJob.ToLogrusFields()).Error("failed to fetch auctions")

				continue
			}

			storeAucsInJobs <- diskstore.WriteAuctionsWithTuplesInJob{
				Tuple:    aucsOutJob.Tuple,
				Auctions: aucsOutJob.AuctionsResponse.Auctions,
			}
			itemIdsInJobs <- aucsOutJob.AuctionsResponse.Auctions.ItemIds()
		}

		close(storeAucsInJobs)
		close(itemIdsInJobs)
	}()

	// spinning up a worker for receiving item-ids from auctions
	go func() {
		results := blizzardv2.ItemIds{}
		for receivedItemIds := range itemIdsInJobs {
			results = results.Merge(receivedItemIds)
		}

		itemIdsOutJob <- results
		close(itemIdsOutJob)
	}()

	// waiting for store-auctions results to drain out
	for storeAucsOutJob := range storeAucsOutJobs {
		if storeAucsOutJob.Err != nil {
			logging.WithFields(storeAucsOutJob.ToLogrusFields()).Error("failed to store auctions")

			return blizzardv2.ItemIds{}, storeAucsOutJob.Err
		}
	}

	// waiting for item-ids to drain out
	receivedItemIds := <-itemIdsOutJob

	return receivedItemIds, nil
}
