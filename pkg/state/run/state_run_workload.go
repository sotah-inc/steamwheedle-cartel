package run

import (
	"log"

	"cloud.google.com/go/storage"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store/regions"
	"github.com/twinj/uuid"
)

type WorkloadStateConfig struct {
	ProjectId string
}

func NewWorkloadState(config WorkloadStateConfig) (WorkloadState, error) {
	// establishing an initial state
	sta := WorkloadState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	// store and storage fields
	sta.IO.StoreClient, err = store.NewClient(config.ProjectId)
	if err != nil {
		log.Fatalf("Failed to create new store client: %s", err.Error())

		return WorkloadState{}, err
	}

	sta.bootBase = store.NewBootBase(sta.IO.StoreClient, regions.USCentral1)
	sta.bootBucket, err = sta.bootBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return WorkloadState{}, err
	}

	sta.realmsBase = store.NewRealmsBase(sta.IO.StoreClient, regions.USCentral1, gameversions.Retail)
	sta.realmsBucket, err = sta.realmsBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return WorkloadState{}, err
	}

	sta.auctionsStoreBase = store.NewAuctionsBaseV2(sta.IO.StoreClient, regions.USCentral1, gameversions.Retail)
	sta.auctionsBucket, err = sta.auctionsStoreBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return WorkloadState{}, err
	}

	sta.auctionManifestStoreBase = store.NewAuctionManifestBaseV2(
		sta.IO.StoreClient,
		regions.USCentral1,
		gameversions.Retail,
	)
	sta.auctionsManifestBucket, err = sta.auctionManifestStoreBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return WorkloadState{}, err
	}

	sta.liveAuctionsStoreBase = store.NewLiveAuctionsBase(sta.IO.StoreClient, regions.USCentral1, gameversions.Retail)
	sta.liveAuctionsBucket, err = sta.liveAuctionsStoreBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return WorkloadState{}, err
	}

	sta.pricelistHistoriesStoreBase = store.NewPricelistHistoriesBaseV2(
		sta.IO.StoreClient,
		regions.USCentral1,
		gameversions.Retail,
	)
	sta.pricelistHistoriesBucket, err = sta.pricelistHistoriesStoreBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return WorkloadState{}, err
	}

	// sotah fields
	sta.regions, err = sta.bootBase.GetRegions(sta.bootBucket)
	if err != nil {
		log.Fatalf("Failed to get regions: %s", err.Error())

		return WorkloadState{}, err
	}

	blizzardCredentials, err := sta.bootBase.GetBlizzardCredentials(sta.bootBucket)
	if err != nil {
		log.Fatalf("Failed to get blizzard-credentials: %s", err.Error())

		return WorkloadState{}, err
	}

	// blizzard fields
	sta.blizzardClient, err = blizzard.NewClient(blizzardCredentials.ClientId, blizzardCredentials.ClientSecret)
	if err != nil {
		log.Fatalf("Failed to create blizzard client: %s", err.Error())

		return WorkloadState{}, err
	}

	return sta, nil
}

type WorkloadState struct {
	state.State

	// store and storage fields
	bootBase   store.BootBase
	bootBucket *storage.BucketHandle

	realmsBase   store.RealmsBase
	realmsBucket *storage.BucketHandle

	auctionsStoreBase store.AuctionsBaseV2
	auctionsBucket    *storage.BucketHandle

	auctionManifestStoreBase store.AuctionManifestBaseV2
	auctionsManifestBucket   *storage.BucketHandle

	liveAuctionsStoreBase store.LiveAuctionsBase
	liveAuctionsBucket    *storage.BucketHandle

	pricelistHistoriesStoreBase store.PricelistHistoriesBaseV2
	pricelistHistoriesBucket    *storage.BucketHandle

	// sotah fields
	regions sotah.RegionList

	// blizzard fields
	blizzardClient blizzard.Client
}
