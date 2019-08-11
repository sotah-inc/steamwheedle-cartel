package prod

import (
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/database"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/hell"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store/regions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
	"github.com/twinj/uuid"
)

type ProdLiveAuctionsStateConfig struct {
	GCloudProjectID string

	MessengerHost string
	MessengerPort int

	LiveAuctionsDatabaseDir string
}

func NewProdLiveAuctionsState(config ProdLiveAuctionsStateConfig) (ProdLiveAuctionsState, error) {
	// establishing an initial state
	liveAuctionsState := ProdLiveAuctionsState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	// connecting to hell
	liveAuctionsState.IO.HellClient, err = hell.NewClient(config.GCloudProjectID)
	if err != nil {
		return ProdLiveAuctionsState{}, err
	}

	// connecting to the messenger host
	liveAuctionsState.IO.Messenger, err = messenger.NewMessenger(config.MessengerHost, config.MessengerPort)
	if err != nil {
		return ProdLiveAuctionsState{}, err
	}

	// initializing a reporter
	liveAuctionsState.IO.Reporter = metric.NewReporter(liveAuctionsState.IO.Messenger)

	// establishing a bus
	logging.Info("Connecting bus-client")
	liveAuctionsState.IO.BusClient, err = bus.NewClient(config.GCloudProjectID, "prod-liveauctions")
	if err != nil {
		return ProdLiveAuctionsState{}, err
	}
	liveAuctionsState.receiveRealmsTopic, err = liveAuctionsState.IO.BusClient.FirmTopic(string(subjects.ReceiveRealms))
	if err != nil {
		log.Fatalf("Failed to get firm topic: %s", err.Error())

		return ProdLiveAuctionsState{}, err
	}

	// establishing a store
	liveAuctionsState.IO.StoreClient, err = store.NewClient(config.GCloudProjectID)
	if err != nil {
		return ProdLiveAuctionsState{}, err
	}

	liveAuctionsState.LiveAuctionsBase = store.NewLiveAuctionsBase(
		liveAuctionsState.IO.StoreClient,
		regions.USCentral1,
		gameversions.Retail,
	)
	liveAuctionsState.LiveAuctionsBucket, err = liveAuctionsState.LiveAuctionsBase.GetFirmBucket()
	if err != nil {
		return ProdLiveAuctionsState{}, err
	}

	bootBase := store.NewBootBase(liveAuctionsState.IO.StoreClient, "us-central1")

	// gathering region-realms
	statuses := sotah.Statuses{}
	bootBucket, err := bootBase.GetFirmBucket()
	if err != nil {
		return ProdLiveAuctionsState{}, err
	}

	regionList, err := bootBase.GetRegions(bootBucket)
	if err != nil {
		return ProdLiveAuctionsState{}, err
	}

	realmsBase := store.NewRealmsBase(liveAuctionsState.IO.StoreClient, "us-central1", gameversions.Retail)
	realmsBucket, err := realmsBase.GetFirmBucket()
	if err != nil {
		return ProdLiveAuctionsState{}, err
	}

	regionRealms := sotah.RegionRealms{}
	for _, region := range regionList {
		realms, err := realmsBase.GetAllRealms(region.Name, realmsBucket)
		if err != nil {
			return ProdLiveAuctionsState{}, err
		}

		regionRealms[region.Name] = realms
	}
	for regionName, realms := range regionRealms {
		statuses[regionName] = sotah.Status{Realms: realms}
	}
	liveAuctionsState.Statuses = statuses

	// ensuring database paths exist
	databasePaths := []string{}
	for regionName, realms := range regionRealms {
		for _, realm := range realms {
			databasePaths = append(databasePaths, fmt.Sprintf(
				"%s/live-auctions/%s/%s",
				config.LiveAuctionsDatabaseDir,
				regionName,
				realm.Slug,
			))
		}
	}
	if err := util.EnsureDirsExist(databasePaths); err != nil {
		return ProdLiveAuctionsState{}, err
	}

	// loading the live-auctions databases
	logging.Info("Connecting to live-auctions databases")
	ladBases, err := database.NewLiveAuctionsDatabases(config.LiveAuctionsDatabaseDir, liveAuctionsState.Statuses)
	if err != nil {
		return ProdLiveAuctionsState{}, err
	}
	liveAuctionsState.IO.Databases.LiveAuctionsDatabases = ladBases

	// establishing bus-listeners
	liveAuctionsState.BusListeners = state.NewBusListeners(state.SubjectBusListeners{
		subjects.ReceiveComputedLiveAuctions: liveAuctionsState.ListenForComputedLiveAuctions,
	})

	// establishing messenger-listeners
	liveAuctionsState.Listeners = state.NewListeners(state.SubjectListeners{
		subjects.Auctions:           liveAuctionsState.ListenForAuctions,
		subjects.OwnersQuery:        liveAuctionsState.ListenForOwnersQuery,
		subjects.PriceList:          liveAuctionsState.ListenForPricelist,
		subjects.OwnersQueryByItems: liveAuctionsState.ListenForOwnersQueryByItems,
	})

	return liveAuctionsState, nil
}

type ProdLiveAuctionsState struct {
	state.State

	Regions  sotah.RegionList
	Statuses sotah.Statuses

	LiveAuctionsBase   store.LiveAuctionsBase
	LiveAuctionsBucket *storage.BucketHandle

	receiveRealmsTopic *pubsub.Topic
}
