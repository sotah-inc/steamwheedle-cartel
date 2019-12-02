package prod

import (
	"fmt"

	"git.sotah.info/steamwheedle-cartel/pkg/state"

	"cloud.google.com/go/storage"
	"git.sotah.info/steamwheedle-cartel/pkg/bus"
	"git.sotah.info/steamwheedle-cartel/pkg/database"
	"git.sotah.info/steamwheedle-cartel/pkg/logging"
	"git.sotah.info/steamwheedle-cartel/pkg/messenger"
	"git.sotah.info/steamwheedle-cartel/pkg/metric"
	"git.sotah.info/steamwheedle-cartel/pkg/sotah"
	"git.sotah.info/steamwheedle-cartel/pkg/sotah/gameversions"
	"git.sotah.info/steamwheedle-cartel/pkg/state/subjects"
	"git.sotah.info/steamwheedle-cartel/pkg/store"
	"git.sotah.info/steamwheedle-cartel/pkg/store/regions"
	"git.sotah.info/steamwheedle-cartel/pkg/util"
	"github.com/twinj/uuid"
)

type ProdPricelistHistoriesStateConfig struct {
	GCloudProjectID string

	MessengerHost string
	MessengerPort int

	PricelistHistoriesDatabaseDir string
}

func NewProdPricelistHistoriesState(config ProdPricelistHistoriesStateConfig) (ProdPricelistHistoriesState, error) {
	// establishing an initial state
	phState := ProdPricelistHistoriesState{
		State: state.NewState(uuid.NewV4(), true),
	}

	// connecting to the messenger host
	mess, err := messenger.NewMessenger(config.MessengerHost, config.MessengerPort)
	if err != nil {
		return ProdPricelistHistoriesState{}, err
	}
	phState.IO.Messenger = mess

	// establishing a bus
	logging.Info("Connecting bus-client")
	busClient, err := bus.NewClient(config.GCloudProjectID, "prod-pricelisthistories")
	if err != nil {
		return ProdPricelistHistoriesState{}, err
	}
	phState.IO.BusClient = busClient

	// establishing a store
	storeClient, err := store.NewClient(config.GCloudProjectID)
	if err != nil {
		return ProdPricelistHistoriesState{}, err
	}
	phState.IO.StoreClient = storeClient

	phState.PricelistHistoriesBase = store.NewPricelistHistoriesBaseV2(
		storeClient,
		regions.USCentral1,
		gameversions.Retail,
	)
	phState.PricelistHistoriesBucket, err = phState.PricelistHistoriesBase.GetFirmBucket()
	if err != nil {
		return ProdPricelistHistoriesState{}, err
	}

	bootBase := store.NewBootBase(storeClient, "us-central1")

	// gathering region-realms
	statuses := sotah.Statuses{}
	bootBucket, err := bootBase.GetFirmBucket()
	if err != nil {
		return ProdPricelistHistoriesState{}, err
	}

	regionList, err := bootBase.GetRegions(bootBucket)
	if err != nil {
		return ProdPricelistHistoriesState{}, err
	}

	realmsBase := store.NewRealmsBase(storeClient, "us-central1", gameversions.Retail)
	realmsBucket, err := realmsBase.GetFirmBucket()
	if err != nil {
		return ProdPricelistHistoriesState{}, err
	}

	regionRealms := sotah.RegionRealms{}
	for _, region := range regionList {
		realms, err := realmsBase.GetAllRealms(region.Name, realmsBucket)
		if err != nil {
			return ProdPricelistHistoriesState{}, err
		}

		regionRealms[region.Name] = realms
	}
	for regionName, realms := range regionRealms {
		statuses[regionName] = sotah.Status{Realms: realms}
	}
	phState.Statuses = statuses

	// ensuring database paths exist
	databasePaths := []string{}
	for regionName, realms := range regionRealms {
		for _, realm := range realms {
			databasePaths = append(databasePaths, fmt.Sprintf(
				"%s/pricelist-histories/%s/%s",
				config.PricelistHistoriesDatabaseDir,
				regionName,
				realm.Slug,
			))
		}
	}
	if err := util.EnsureDirsExist(databasePaths); err != nil {
		return ProdPricelistHistoriesState{}, err
	}

	// initializing a reporter
	phState.IO.Reporter = metric.NewReporter(mess)

	// loading the pricelist-histories databases
	logging.Info("Connecting to pricelist-histories databases")
	phdBases, err := database.NewPricelistHistoryDatabases(config.PricelistHistoriesDatabaseDir, phState.Statuses)
	if err != nil {
		return ProdPricelistHistoriesState{}, err
	}
	phState.IO.Databases.PricelistHistoryDatabases = phdBases

	// loading the meta database
	logging.Info("Connecting to the meta database")
	metaDatabase, err := database.NewMetaDatabase(config.PricelistHistoriesDatabaseDir)
	if err != nil {
		return ProdPricelistHistoriesState{}, err
	}
	phState.IO.Databases.MetaDatabase = metaDatabase

	// establishing bus-listeners
	phState.BusListeners = state.NewBusListeners(state.SubjectBusListeners{
		subjects.ReceiveComputedPricelistHistories: phState.ListenForComputedPricelistHistories,
	})

	// establishing messenger-listeners
	phState.Listeners = state.NewListeners(state.SubjectListeners{
		subjects.PriceListHistory: phState.ListenForPriceListHistory,
	})

	return phState, nil
}

type ProdPricelistHistoriesState struct {
	state.State

	Statuses sotah.Statuses

	PricelistHistoriesBase   store.PricelistHistoriesBaseV2
	PricelistHistoriesBucket *storage.BucketHandle
}
