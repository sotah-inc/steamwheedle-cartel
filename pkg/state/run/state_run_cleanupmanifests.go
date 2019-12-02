package run

import (
	"log"

	"cloud.google.com/go/storage"
	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store/regions"
)

type CleanupManifestsStateConfig struct {
	ProjectId string
}

func NewCleanupManifestsState(config CleanupManifestsStateConfig) (CleanupManifestsState, error) {
	// establishing an initial state
	sta := CleanupManifestsState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	sta.IO.StoreClient, err = store.NewClient(config.ProjectId)
	if err != nil {
		log.Fatalf("Failed to create new store client: %s", err.Error())

		return CleanupManifestsState{}, err
	}

	sta.manifestBase = store.NewAuctionManifestBaseV2(sta.IO.StoreClient, regions.USCentral1, gameversions.Retail)
	sta.manifestBucket, err = sta.manifestBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm raw-auctions bucket: %s", err.Error())

		return CleanupManifestsState{}, err
	}

	return sta, nil
}

type CleanupManifestsState struct {
	state.State

	manifestBase   store.AuctionManifestBaseV2
	manifestBucket *storage.BucketHandle
}
