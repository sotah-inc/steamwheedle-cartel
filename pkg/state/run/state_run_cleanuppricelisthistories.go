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

type CleanupPricelistHistoriesStateConfig struct {
	ProjectId string
}

func NewCleanupPricelistHistoriesState(
	config CleanupPricelistHistoriesStateConfig,
) (CleanupPricelistHistoriesState, error) {
	// establishing an initial state
	sta := CleanupPricelistHistoriesState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	sta.IO.StoreClient, err = store.NewClient(config.ProjectId)
	if err != nil {
		log.Fatalf("Failed to create new store client: %s", err.Error())

		return CleanupPricelistHistoriesState{}, err
	}

	sta.pricelistHistoriesBase = store.NewPricelistHistoriesBaseV2(
		sta.IO.StoreClient,
		regions.USCentral1,
		gameversions.Retail,
	)
	sta.pricelistHistoriesBucket, err = sta.pricelistHistoriesBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return CleanupPricelistHistoriesState{}, err
	}

	return sta, nil
}

type CleanupPricelistHistoriesState struct {
	state.State

	pricelistHistoriesBase   store.PricelistHistoriesBaseV2
	pricelistHistoriesBucket *storage.BucketHandle
}
