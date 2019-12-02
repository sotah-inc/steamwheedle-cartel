package run

import (
	"log"

	"cloud.google.com/go/storage"
	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/store"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/store/regions"
)

type CleanupAuctionsStateConfig struct {
	ProjectId string
}

func NewCleanupAuctionsState(config CleanupAuctionsStateConfig) (CleanupAuctionsState, error) {
	// establishing an initial state
	sta := CleanupAuctionsState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	sta.IO.StoreClient, err = store.NewClient(config.ProjectId)
	if err != nil {
		log.Fatalf("Failed to create new store client: %s", err.Error())

		return CleanupAuctionsState{}, err
	}

	sta.auctionsBase = store.NewAuctionsBaseV2(sta.IO.StoreClient, regions.USCentral1, gameversions.Retail)
	sta.auctionsBucket, err = sta.auctionsBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm raw-auctions bucket: %s", err.Error())

		return CleanupAuctionsState{}, err
	}

	return sta, nil
}

type CleanupAuctionsState struct {
	state.State

	auctionsBase   store.AuctionsBaseV2
	auctionsBucket *storage.BucketHandle
}
