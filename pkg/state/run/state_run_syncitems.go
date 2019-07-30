package run

import (
	"log"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store"
	"github.com/twinj/uuid"
)

type SyncItemsStateConfig struct {
	ProjectId string
}

func NewSyncItemsState(config SyncItemsStateConfig) (SyncItemsState, error) {
	// establishing an initial state
	sta := SyncItemsState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	// initializing a bus client
	sta.IO.BusClient, err = bus.NewClient(config.ProjectId, "run-sync-items")
	if err != nil {
		log.Fatalf("Failed to create new bus client: %s", err.Error())

		return SyncItemsState{}, err
	}
	sta.receiveSyncedItemsTopic, err = sta.IO.BusClient.FirmTopic(string(subjects.ReceiveSyncedItems))
	if err != nil {
		log.Fatalf("Failed to get firm topic: %s", err.Error())

		return SyncItemsState{}, err
	}

	// initializing a store client
	sta.IO.StoreClient, err = store.NewClient(config.ProjectId)
	if err != nil {
		log.Fatalf("Failed to create new store client: %s", err.Error())

		return SyncItemsState{}, err
	}

	return sta, nil
}

type SyncItemsState struct {
	state.State

	receiveSyncedItemsTopic *pubsub.Topic

	itemsBase   store.ItemsBase
	itemsBucket *storage.BucketHandle

	blizzardClient blizzard.Client

	primaryRegion sotah.Region
}
