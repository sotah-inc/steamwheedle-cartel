package prod

import (
	"cloud.google.com/go/storage"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/database"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store/regions"
	"github.com/twinj/uuid"
)

type ItemsStateConfig struct {
	GCloudProjectID string

	MessengerHost string
	MessengerPort int

	ItemsDatabaseDir string
}

func NewProdItemsState(config ItemsStateConfig) (ItemsState, error) {
	// establishing an initial state
	itemsState := ItemsState{
		State: state.NewState(uuid.NewV4(), true),
	}

	// connecting to the messenger host
	mess, err := messenger.NewMessenger(config.MessengerHost, config.MessengerPort)
	if err != nil {
		return ItemsState{}, err
	}
	itemsState.IO.Messenger = mess

	// establishing a bus
	logging.Info("Connecting bus-client")
	busClient, err := bus.NewClient(config.GCloudProjectID, "prod-items")
	if err != nil {
		return ItemsState{}, err
	}
	itemsState.IO.BusClient = busClient

	// establishing a store
	storeClient, err := store.NewClient(config.GCloudProjectID)
	if err != nil {
		return ItemsState{}, err
	}
	itemsState.IO.StoreClient = storeClient

	itemsState.ItemsBase = store.NewItemsBase(storeClient, regions.USCentral1, gameversions.Retail)
	itemsState.ItemsBucket, err = itemsState.ItemsBase.GetFirmBucket()
	if err != nil {
		return ItemsState{}, err
	}

	// initializing a reporter
	itemsState.IO.Reporter = metric.NewReporter(mess)

	// loading the items database
	logging.Info("Connecting to items database")
	iBase, err := database.NewItemsDatabase(config.ItemsDatabaseDir)
	if err != nil {
		return ItemsState{}, err
	}
	itemsState.IO.Databases.ItemsDatabase = iBase

	// establishing bus-listeners
	itemsState.BusListeners = state.NewBusListeners(state.SubjectBusListeners{
		subjects.FilterInItemsToSync: itemsState.ListenForFilterIn,
		subjects.ReceiveSyncedItems:  itemsState.ListenForSyncedItems,
	})

	// establishing messenger-listeners
	itemsState.Listeners = state.NewListeners(state.SubjectListeners{
		subjects.Items:      itemsState.ListenForItems,
		subjects.ItemsQuery: itemsState.ListenForItemsQuery,
	})

	return itemsState, nil
}

type ItemsState struct {
	state.State

	ItemsBase   store.ItemsBase
	ItemsBucket *storage.BucketHandle
}
