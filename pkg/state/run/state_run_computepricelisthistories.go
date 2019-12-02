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

type ComputePricelistHistoriesStateConfig struct {
	ProjectId string
}

func NewComputePricelistHistoriesState(
	config ComputePricelistHistoriesStateConfig,
) (ComputePricelistHistoriesState, error) {
	// establishing an initial state
	sta := ComputePricelistHistoriesState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	sta.IO.StoreClient, err = store.NewClient(config.ProjectId)
	if err != nil {
		log.Fatalf("Failed to create new store client: %s", err.Error())

		return ComputePricelistHistoriesState{}, err
	}

	sta.auctionsStoreBase = store.NewAuctionsBaseV2(sta.IO.StoreClient, regions.USCentral1, gameversions.Retail)
	sta.auctionsBucket, err = sta.auctionsStoreBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return ComputePricelistHistoriesState{}, err
	}

	sta.pricelistHistoriesStoreBase = store.NewPricelistHistoriesBaseV2(
		sta.IO.StoreClient,
		regions.USCentral1,
		gameversions.Retail,
	)
	sta.pricelistHistoriesBucket, err = sta.pricelistHistoriesStoreBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return ComputePricelistHistoriesState{}, err
	}

	return sta, nil
}

type ComputePricelistHistoriesState struct {
	state.State

	auctionsStoreBase store.AuctionsBaseV2
	auctionsBucket    *storage.BucketHandle

	pricelistHistoriesStoreBase store.PricelistHistoriesBaseV2
	pricelistHistoriesBucket    *storage.BucketHandle
}
