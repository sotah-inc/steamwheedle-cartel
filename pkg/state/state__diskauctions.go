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

func (sta DiskAuctionsState) Collect(tuples []blizzardv2.RegionConnectedRealmTuple) error {
	// spinning up workers
	aucsOutJobs := sta.BlizzardState.ResolveAuctions(tuples)
	storeAucsInJobs := make(chan diskstore.WriteAuctionsWithTuplesInJob)
	storeAucsOutJobs := sta.DiskStore.WriteAuctionsWithTuples(storeAucsInJobs)

	// interpolating auctions-out jobs into store-auctions-in jobs
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
		}

		close(storeAucsInJobs)
	}()

	// waiting for results to drain out
	for storeAucsOutJob := range storeAucsOutJobs {
		if storeAucsOutJob.Err != nil {
			logging.WithFields(storeAucsOutJob.ToLogrusFields()).Error("failed to store auctions")

			return storeAucsOutJob.Err
		}
	}

	return nil
}
