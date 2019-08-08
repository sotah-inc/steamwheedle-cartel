package run

import (
	"log"

	"cloud.google.com/go/storage"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store/regions"
	"github.com/twinj/uuid"
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
