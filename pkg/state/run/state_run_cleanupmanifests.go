package run

import (
	"log"

	"cloud.google.com/go/storage"
	"git.sotah.info/steamwheedle-cartel/pkg/sotah/gameversions"
	"git.sotah.info/steamwheedle-cartel/pkg/state"
	"git.sotah.info/steamwheedle-cartel/pkg/store"
	"git.sotah.info/steamwheedle-cartel/pkg/store/regions"
	"github.com/twinj/uuid"
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
